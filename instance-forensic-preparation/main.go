package main

import (
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"time"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
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
func HandleRequest(instanceId string) (error) {
	log.Logger = zerolog.New(os.Stdout).With().Caller().Logger()

	if instanceId == "" {
		return errors.New("Empty InstanceId")
	}

	forensic := &EC2Forensic{
		svc: ec2.New(session.Must(session.NewSession())),
		VpcId: forensicVpcId,
		SubnetId: forensicSubnetId,
		SecurityGroupId: forensicSgId,
		InstanceId: instanceId,
	}

	// Stop Instance (Immediately, because it is suspected to be infected)
	log.Info().Str("duration", returnDuration()).Str("status", "stop instance").Msg("started")
	if err := forensic.StopInstance(); err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "stop instance").Msg("failed")
	}
	log.Info().Str("duration", returnDuration()).Str("status", "stop instance").Msg("succeeded")

	// Create a snapshot for an evidence
	log.Info().Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("started")
	snapShotId, err := forensic.CreateEvidenceSnapshot()
	if err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("failed")
	}
	log.Info().Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("succeeded")

	// Create EBS from snapshot
	log.Info().Str("duration", returnDuration()).Str("status", "create an EBS volume").Msg("started")
	volumeId, err := forensic.CreateEvidenceEBS(snapShotId)
	if err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "create an EBS volume").Msg("failed")
	}
	log.Info().Str("duration", returnDuration()).Str("status", "create an EBS volume").Msg("succeeded")

	// Start up a forensic workstation
	log.Info().Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("started")
	workstationId, err := forensic.StartForensicWorkstation()
	if err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("failed")
	}
	log.Info().Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("succeeded")

	// Attach Evidence EBS to Workstation
	log.Info().Str("duration", returnDuration()).Str("status", "Attaching Volume").Msg("started")
	if err := forensic.AttachEvidenceToWorkstation(workstationId, volumeId); err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "Attaching Volume").Msg("failed")
	}
	log.Info().Str("duration", returnDuration()).Str("status", "Attaching Volume").Msg("succeeded")

	return nil
}

type EC2Forensic struct {
	svc *ec2.EC2
	InstanceId string
	VpcId string
	SubnetId string
	SecurityGroupId string
}

// StopInstance stops running instance.
// TODO Check status of instance first: exists? just stopped?
func (e *EC2Forensic) StopInstance() error {
	log.Info().Str("tagetInstanceId", e.InstanceId)
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
		Attribute: aws.String("blockDeviceMapping"),
		InstanceId: aws.String(e.InstanceId),
	}

	createSnapshotInput := &ec2.CreateSnapshotInput{
		Description: aws.String("Snapshot taken for forensic purpose"),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("snapshot"),
				Tags: []*ec2.Tag{{Key: aws.String("Name"), Value:aws.String("forensic-snapshot")}},
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
	for snapShotState != "completed"  {
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
		AvailabilityZone: aws.String("ap-northeast-1d"), //ToDO Dynamically retrieve from forensic_subnet-id, config, or whatever
		SnapshotId: aws.String(snapshotId),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("volume"),
				Tags: []*ec2.Tag{{Key: aws.String("Name"), Value:aws.String("forensic-ebs-volume")}},
			},
		},
	}

	output, err := e.svc.CreateVolume(input)
	if err != nil {
		return "", err
	}

	//TODO Implement the following
	//e.svc.DescribeVolumeStatus()

	return *output.VolumeId, nil
}

//TODO put into config // maybe t2.large is better according to https://www.sans.org/reading-room/whitepapers/cloud/digital-forensic-analysis-amazon-linux-ec2-instances-38235?
//ToDo Add User Data or make it the latest
func (e *EC2Forensic) StartForensicWorkstation() (workstationId string, err error) {
	input := &ec2.RunInstancesInput{
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{ Key: aws.String("Name"), Value:aws.String("forensic-workstation")},
					{ Key: aws.String("Target"), Value: aws.String(e.InstanceId)},
				},
			},
		},
		MaxCount: aws.Int64(1),
		MinCount: aws.Int64(1),
		SecurityGroupIds: []*string{aws.String(forensicSgId)},
		SubnetId: aws.String(forensicSubnetId),
		InstanceType: aws.String("t2.micro"), // maybe t2.large
		ImageId: aws.String("ami-e99f4896"),  // maybe Ubuntu Server 16.04 LTS (HVM), SSD Volume Type - ami-940cdceb (2018/07/15)
	}

	output, err := e.svc.RunInstances(input)
	if err != nil {
		return "", err
	}

	var instanceState string
	for instanceState != "running" {
		output, err := e.svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
			InstanceIds:[]*string{output.Instances[0].InstanceId},
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
		VolumeId: aws.String(evidenceVolumeId),
		Device: aws.String("/dev/sdf"),
	})
	return
}

// CloudWatchEventForGuardDuty: https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_findings_cloudwatch.html
// ToDo slack-notificationのmain.goとあわせる
type CloudWatchEventForGuardDuty struct {
	Account     string           `json:"account"`
	//Detail      GuardDutyFinding `json:"detail"`
	Detail_type string           `json:"detail-type"`
	ID          string           `json:"id"`
	Region      string           `json:"region"`
	Resources   []interface{}    `json:"resources"`
	Source      string           `json:"source"`
	Time        string           `json:"time"`
	Version     string           `json:"version"`
}

func returnDuration() string {
	return string(time.Now().Format("05.00"))
}

