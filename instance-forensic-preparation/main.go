package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/nlopes/slack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"time"
)

var slackURL, forensicVpcId, forensicSubnetId, forensicSgId string

func init() {
	slackURL = os.Getenv("SLACK_URL")
	forensicVpcId = os.Getenv("FORENSIC_VPC_ID")
	forensicSubnetId = os.Getenv("FORENSIC_SUBNET_ID")
	forensicSgId = os.Getenv("FORENSIC_SG_ID")
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

//TODO Check existence of ForensicVPC/ForensicSubnet
//TODO Add Date
func main() {
	if slackURL == "" || forensicVpcId == "" || forensicSubnetId == "" || forensicSgId == "" {
		log.Fatal().Msg("you must set Env Var `SLACK_URL`, `FORENSIC_VPC_ID`, `FORENSIC_SG_ID` and `FORENSIC_SUBNET_ID`")
	}
	lambda.Start(HandleRequest)
}

//func HandleRequest(request CloudWatchEventForGuardDuty) (string, error) {
func HandleRequest(instanceId string) error {
	log.Logger = zerolog.New(os.Stdout).With().Caller().Logger()

	if instanceId == "" {
		return errors.New("empty instanceId")
	}

	forensic := &EC2Forensic{
		svc:             ec2.New(session.Must(session.NewSession())),
		VpcId:           forensicVpcId,
		SubnetId:        forensicSubnetId,
		SecurityGroupId: forensicSgId,
		InstanceId:      instanceId,
	}

	notify(false, false, nil, fmt.Sprintf("Started forensic preparation: %v", instanceId))

	// Stop Instance (Immediately, because it is suspected to be infected)
	if err := forensic.StopInstance(); err != nil {
		notify(true, false, nil, "Failed Stopping Instance")
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "stop instance").Msg("failed")
	}
	notify(false, false, nil, fmt.Sprintf("Stopped Instance: %v", instanceId))

	// Create a snapshot for an evidence
	snapShotId, err := forensic.CreateEvidenceSnapshot()
	if err != nil {
		notify(true, false, nil, "Failed taking a Snapshot")
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("failed")
	}
	notify(false, false, nil, fmt.Sprintf("Created a snapshot for evidence: %v", snapShotId))

	// Create EBS from snapshot
	volumeId, err := forensic.CreateEvidenceEBS(snapShotId)
	if err != nil {
		notify(true, false, nil, "Failed creating EBS from snapshot")
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "create an EBS volume").Msg("failed")
	}
	notify(false, false, nil, fmt.Sprintf("Created EBS volume from snapshot: %v", volumeId))

	// Start up a forensic workstation
	// TODO This can be run concurrently
	workstationId, err := forensic.StartForensicWorkstation()
	if err != nil {
		notify(true, false, nil, "Failed starting up forensic workstation")
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("failed")
	}
	notify(false, false, nil, fmt.Sprintf("Started up forensic workstation: %v", workstationId))

	// Attach Evidence EBS to Workstation
	if err := forensic.AttachEvidenceToWorkstation(workstationId, volumeId); err != nil {
		notify(true, false, nil, "Failed attaching the evidence volume to forensic instance")
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "Attaching Volume").Msg("failed")
	}
	notify(false, false, nil, fmt.Sprintf("Attached the evidence volume (%v) to forensic instance (%v)", volumeId, workstationId))

	notify(false, true, nil, "Finished preparation for forensic")
	return nil
}

type EC2Forensic struct {
	svc             *ec2.EC2
	InstanceId      string
	VpcId           string
	SubnetId        string
	SecurityGroupId string
}

// StopInstance stops running instance.
// TODO Check status of instance first: exists? just stopped?
func (e *EC2Forensic) StopInstance() error {
	var stopInstancesInput = &ec2.StopInstancesInput{
		Force:       aws.Bool(true),
		InstanceIds: []*string{aws.String(e.InstanceId)},
	}
	var describeInstancesInput = &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(e.InstanceId)},
	}

	if _, err := e.svc.StopInstances(stopInstancesInput); err != nil {
		return err
	}
	if err := e.svc.WaitUntilInstanceStopped(describeInstancesInput); err != nil {
		return err
	}

	return nil
}

// Take SnapShot of suspected EC2 instance for taking evidence
func (e *EC2Forensic) CreateEvidenceSnapshot() (snapshotId string, err error) {
	describeEc2AttributeInput := &ec2.DescribeInstanceAttributeInput{
		Attribute:  aws.String("blockDeviceMapping"),
		InstanceId: aws.String(e.InstanceId),
	}

	createSnapshotInput := &ec2.CreateSnapshotInput{
		Description: aws.String("Snapshot taken for forensic purpose"),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("snapshot"),
				Tags:         []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String("forensic-snapshot")}},
			},
		},
		//Encrypted: aws.Bool(true), // for now
	}

	output, err := e.svc.DescribeInstanceAttribute(describeEc2AttributeInput)
	if err != nil {
		return "", err
	}

	createSnapshotInput.VolumeId = output.BlockDeviceMappings[0].Ebs.VolumeId
	snapShot, err := e.svc.CreateSnapshot(createSnapshotInput)
	if err != nil {
		return "", err
	}
	// Check State
	var snapShotState string
	for snapShotState != "completed" {
		output, err := e.svc.DescribeSnapshots(&ec2.DescribeSnapshotsInput{SnapshotIds: []*string{snapShot.SnapshotId}})
		if err != nil {
			return "", nil
		}
		snapShotState = *output.Snapshots[0].State
	}

	return *snapShot.SnapshotId, nil
}

func (e *EC2Forensic) CreateEvidenceEBS(snapshotId string) (volumeId string, err error) {
	input := &ec2.CreateVolumeInput{
		//ToDO Dynamically retrieve from forensic_subnet-id, config, or whatever
		AvailabilityZone: aws.String("ap-northeast-1d"),
		SnapshotId:       aws.String(snapshotId),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("volume"),
				Tags:         []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String("forensic-ebs-volume")}},
			},
		},
	}

	output, err := e.svc.CreateVolume(input)
	if err != nil {
		return "", err
	}
	// Check State
	var volumeState string
	if volumeState != "ok" {
		output, err := e.svc.DescribeVolumeStatus(&ec2.DescribeVolumeStatusInput{
			VolumeIds: []*string{output.VolumeId},
		})
		if err != nil {
			return "", err
		}
		volumeState = *output.VolumeStatuses[0].VolumeStatus.Status
	}

	return *output.VolumeId, nil
}

//ToDo Add User Data and make it the latest
func (e *EC2Forensic) StartForensicWorkstation() (workstationId string, err error) {
	input := &ec2.RunInstancesInput{
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{Key: aws.String("Name"), Value: aws.String("forensic-workstation")},
					{Key: aws.String("Target"), Value: aws.String(e.InstanceId)},
				},
			},
		},
		MaxCount:         aws.Int64(1),
		MinCount:         aws.Int64(1),
		SecurityGroupIds: []*string{aws.String(forensicSgId)},
		SubnetId:         aws.String(forensicSubnetId),
		// Recommended in https://www.sans.org/reading-room/whitepapers/cloud/digital-forensic-analysis-amazon-linux-ec2-instances-38235?
		InstanceType: aws.String("t2.large"),
		// Latest(2018/07/15) ami id of Ubuntu Server 16.04 LTS (HVM), SSD Volume Type
		// Recommended in https://www.sans.org/reading-room/whitepapers/cloud/digital-forensic-analysis-amazon-linux-ec2-instances-38235?
		ImageId: aws.String("ami-940cdceb"),
	}

	output, err := e.svc.RunInstances(input)
	if err != nil {
		return "", err
	}

	var instanceState string
	for instanceState != "running" {
		output, err := e.svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
			InstanceIds:         []*string{output.Instances[0].InstanceId},
			IncludeAllInstances: aws.Bool(true),
		})
		if err != nil {
			return "", err
		}
		instanceState = *output.InstanceStatuses[0].InstanceState.Name

	}

	return *output.Instances[0].InstanceId, nil
}

// TODO check if /dev/sdf is not taken
func (e *EC2Forensic) AttachEvidenceToWorkstation(workstationId, evidenceVolumeId string) (err error) {
	_, err = e.svc.AttachVolume(&ec2.AttachVolumeInput{
		InstanceId: aws.String(workstationId),
		VolumeId:   aws.String(evidenceVolumeId),
		Device:     aws.String("/dev/sdf"),
	})
	return
}

func returnDuration() string {
	return string(time.Now().Format("05.00"))
}

func notify(isFailed bool, isCompleted bool, err error, message string) {
	color := "#707070"
	if isFailed {
		color = "warning"
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", message).Msg("failed")
	} else if isCompleted {
		color = "good"
		log.Info().Str("duration", returnDuration()).Str("status", message).Msg("completed")
	} else {
		log.Info().Str("duration", returnDuration()).Str("status", message).Msg("started")
	}

	payload := slack.PostMessageParameters{
		Attachments: []slack.Attachment{{
			Color:   color,
			Pretext: "Forensic Preparation Status",
			Title:   message,
		}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Info().Str("status", "failed").Msg("slack notification failed: failed to encode payload")
	}

	if _, err := http.Post(slackURL, "application/json", bytes.NewReader(body)); err != nil {
		log.Info().Str("status", "failed").Msg("slack notification failed: http post failed")
	}
}
