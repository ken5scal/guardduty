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

var timeout = time.Duration(5 * time.Minute)

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

	//awsInstanceId := []*string{aws.String(instanceId)}
	//stopInstancesInput.InstanceIds = awsInstanceId
	//describeInstancesInput.InstanceIds = awsInstanceId
	//describeEc2AttributeInput.InstanceId = aws.String(instanceId)

	//s := session.Must(session.NewSession())
	//svc := ec2.New(s)

	// Stop Instance
	log.Info().Str("duration", returnDuration()).Str("status", "stop instance").Msg("started")
	forensic.StopInstance()
	log.Info().Str("duration", returnDuration()).Str("status", "stop instance").Msg("succeeded")


	//log.Info().Str("duration", returnDuration()).Str("status", "stop instance").Msg("started")
	//if _, err := svc.StopInstances(stopInstancesInput); err != nil {
	//	log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "stopping instance").Msg("failed")
	//} else {
	//	if err := svc.WaitUntilInstanceStopped(describeInstancesInput); err != nil {
	//		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "stopping instance").Msg("failed")
	//	}
	//}
	//log.Info().Str("duration", returnDuration()).Str("status", "stop instance").Msg("succeeded")

	// Describe Instance
	//log.Info().Str("duration", returnDuration()).Str("status", "describe instance").Msg("started")
	//out, err := svc.DescribeInstanceAttribute(describeEc2AttributeInput)
	//if err != nil {
	//	log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "describe instance").Msg("failed")
	//}
	//log.Info().Str("duration", returnDuration()).Str("status", "describe instance").Msg("succeeded")

	/*
	//  Create a Snapshot
	 */
	// assuming there is only one volume
	//createSnapshotInput.VolumeId = out.BlockDeviceMappings[0].Ebs.VolumeId
	//createSnapshotInput.TagSpecifications = []*ec2.TagSpecification{
	//	{
	//		ResourceType: aws.String("snapshot"),
	//		Tags: []*ec2.Tag{{Key: aws.String("Name"), Value:aws.String("forensic-snapshot")}},
	//	},
	//}
	log.Info().Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("started")
	snapShotId, err := forensic.TakeSnapshot()
	if err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("failed")
	}
	//snapShot, err := svc.CreateSnapshotWithContext(ctx, createSnapshotInput)
	//if err != nil {
	//	log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("failed")
	//}
	//// Check State
	//var snapShotState string
	//for snapShotState != "completed"  {
	//	dso, err := svc.DescribeSnapshots(
	//		&ec2.DescribeSnapshotsInput{SnapshotIds: []*string{snapShot.SnapshotId}})
	//	if err != nil {
	//		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "create an AMI").Msg("failed")
	//	}
	//	snapShotState = *dso.Snapshots[0].State
	//}
	log.Info().Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("succeeded")

	// Create EBS
	log.Info().Str("duration", returnDuration()).Str("status", "create an EBS volume").Msg("started")
	volumeId, err := forensic.CreateBackupEBS(snapShotId)
	//cvi := &ec2.CreateVolumeInput{
	//	//ToDO Dynamically retrieve from forensic_subnet-id, config, or whatever
	//	AvailabilityZone: aws.String("ap-northeast-1d"),
	//	SnapshotId: snapShot.SnapshotId,
	//	TagSpecifications: []*ec2.TagSpecification{
	//		{
	//			ResourceType: aws.String("volume"),
	//			Tags: []*ec2.Tag{
	//				{
	//					Key: aws.String("Name"), Value:aws.String("forensic-ebs-volume"),
	//				},
	//			},
	//		},
	//	},
	//}
	//cvo, err := svc.CreateVolume(cvi)
	if err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "create an EBS volume").Msg("failed")
	}
	log.Info().Str("duration", returnDuration()).Str("status", "create an EBS volume").Msg("succeeded")

	// Run Instance in Forensic VPC
	log.Info().Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("started")
	//ro := &ec2.RunInstancesInput{
	//	TagSpecifications: []*ec2.TagSpecification{
	//		{
	//			ResourceType: aws.String("instance"),
	//			Tags: []*ec2.Tag{
	//				{
	//					Key: aws.String("Name"), Value:aws.String("forensic-instance"),
	//				},
	//			},
	//		},
	//	},
	//	MaxCount: aws.Int64(1),
	//	MinCount: aws.Int64(1),
	//	SecurityGroupIds: []*string{aws.String(forensicSgId)},
	//	SubnetId: aws.String(forensicSubnetId),
	//	InstanceType: aws.String("t2.micro"), //TODO put into config // maybe t2.large is better according to https://www.sans.org/reading-room/whitepapers/cloud/digital-forensic-analysis-amazon-linux-ec2-instances-38235?
	//	ImageId: aws.String("ami-e99f4896"), //ToDo Add User Data or make it the latest
	//}
	//
	//re, err := svc.RunInstances(ro)
	//if err != nil {
	//	log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("failed")
	//}
	//
	//var isRunning bool
	//for !isRunning {
	//	diso, err := svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
	//		InstanceIds:[]*string{re.Instances[0].InstanceId},
	//		IncludeAllInstances: aws.Bool(true),
	//	})
	//	if err != nil {
	//		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("failed")
	//	}
	//	isRunning = (*diso.InstanceStatuses[0].InstanceState.Name == "running")
	//}
	workstationId, err := forensic.StartForensicWorkstation()
	if err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("failed")
	}
	log.Info().Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("succeeded")

	log.Info().Str("duration", returnDuration()).Str("status", "Attaching Volume").Msg("started")
	if err := forensic.AttachEvidenceToWorkstation(workstationId, volumeId); err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "Attaching Volume").Msg("failed")
	}
	//if _, err := svc.AttachVolume(&ec2.AttachVolumeInput{
	//	InstanceId: re.Instances[0].InstanceId,
	//	VolumeId: cvo.VolumeId,
	//	Device: aws.String("/dev/sdf"), //TODO Change device name based on already running instance.
	//}); err != nil {
	//	log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "Attaching Volume").Msg("failed")
	//}
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
	var stopInstancesInput = &ec2.StopInstancesInput{
		Force:       aws.Bool(true),
		InstanceIds: []*string{},
	}
	var describeInstancesInput = &ec2.DescribeInstancesInput{
		InstanceIds: []*string{},
	}

	if _, err := e.svc.StopInstances(stopInstancesInput); err != nil {
		return err
	}
	if err := e.svc.WaitUntilInstanceStopped(describeInstancesInput); err != nil {
		return err
	}

	return nil
}

// Take SnapShot of EC2 instance
func (e *EC2Forensic) TakeSnapshot() (snapshotId string, err error) {
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

func (e *EC2Forensic) CreateBackupEBS(snapshotId string) (volumeId string, err error) {
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
	for instanceState == "running" {
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

