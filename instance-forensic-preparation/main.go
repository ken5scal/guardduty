package main

import (
	"context"
	"errors"
	"fmt"
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
var stopInstancesInput = &ec2.StopInstancesInput{
	Force:       aws.Bool(true),
	InstanceIds: []*string{},
}
var describeInstancesInput = &ec2.DescribeInstancesInput{
	InstanceIds: []*string{},
}
var createSnapshotInput = &ec2.CreateSnapshotInput{
	Description: aws.String("Snapshot taken for forensic purpose"),
	TagSpecifications: []*ec2.TagSpecification{},
	//Encrypted: aws.Bool(true), // for now
}
var describeEc2AttributeInput = &ec2.DescribeInstanceAttributeInput{
	Attribute: aws.String("blockDeviceMapping"),
}

var slackURL, forensicVpcId, forensicSubnetId, forensicSgId string

func init() {
	slackURL = os.Getenv("SLACK_URL")
	forensicVpcId = os.Getenv("FORENSIC_VPC_ID")
	forensicSubnetId = os.Getenv("FORENSIC_SUBNET_ID")
	forensicSgId = os.Getenv("FORENSIC_SG_ID")
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	if slackURL == "" || forensicVpcId == "" || forensicSubnetId == "" || forensicSgId == "" {
		log.Fatal().Msg("you must set Env Var `SLACK_URL`, `FORENSIC_VPC_ID`, `FORENSIC_SG_ID` and `FORENSIC_SUBNET_ID`")
	}
	//TODO Check existence of ForensicVPC/ForensicSubnet
	lambda.Start(HandleRequest)
}
//func HandleRequest(request CloudWatchEventForGuardDuty) (string, error) {
func HandleRequest(instanceId string) (string, error) {
	log.Logger = zerolog.New(os.Stdout).With().Caller().Logger()//.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	if instanceId == "" {
		return "", errors.New("Empty InstanceId")
	}

	awsInstanceId := []*string{aws.String(instanceId)}
	stopInstancesInput.InstanceIds = awsInstanceId
	describeInstancesInput.InstanceIds = awsInstanceId
	describeEc2AttributeInput.InstanceId = aws.String(instanceId)

	s := session.Must(session.NewSession())
	svc := ec2.New(s)

	// Abort the upload if it takes more than the passed in timeout.
	ctx, cancelFn := context.WithTimeout(context.Background(), 5 *time.Minute)
	defer cancelFn()

	// Stop Instance
	// TODO Check status of instance first: exists? just stopped?
	log.Info().Str("duration", returnDuration()).Str("status", "stop instance").Msg("started")
	if _, err := svc.StopInstancesWithContext(ctx, stopInstancesInput); err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "stopping instance").Msg("failed")
	} else {
		if err := svc.WaitUntilInstanceStoppedWithContext(ctx, describeInstancesInput); err != nil {
			log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "stopping instance").Msg("failed")
		}
	}
	log.Info().Str("duration", returnDuration()).Str("status", "stop instance").Msg("succeeded")

	// Describe Instance
	log.Info().Str("duration", returnDuration()).Str("status", "describe instance").Msg("started")
	out, err := svc.DescribeInstanceAttributeWithContext(ctx, describeEc2AttributeInput)
	if err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "describe instance").Msg("failed")
	}
	log.Info().Str("duration", returnDuration()).Str("status", "describe instance").Msg("succeeded")

	/*
	//  Create a Snapshot
	 */
	// assuming there is only one volume
	createSnapshotInput.VolumeId = out.BlockDeviceMappings[0].Ebs.VolumeId
	createSnapshotInput.TagSpecifications = []*ec2.TagSpecification{
		{
			ResourceType: aws.String("snapshot"),
			Tags: []*ec2.Tag{{Key: aws.String("Name"), Value:aws.String("forensic-snapshot")}},
		},
	}
	log.Info().Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("started")
	snapShot, err := svc.CreateSnapshotWithContext(ctx, createSnapshotInput)
	if err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("failed")
	}
	// Check State
	var snapShotState string
	for snapShotState != "completed"  {
		dso, err := svc.DescribeSnapshots(
			&ec2.DescribeSnapshotsInput{SnapshotIds: []*string{snapShot.SnapshotId}})
		if err != nil {
			log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "create an AMI").Msg("failed")
		}
		snapShotState = *dso.Snapshots[0].State
	}
	log.Info().Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("succeeded")

	// Create EBS/AMI
	log.Info().Str("duration", returnDuration()).Str("status", "create an EBS volume").Msg("started")
	//ii := &ec2.ImportImageInput{
	//	DiskContainers: []*ec2.ImageDiskContainer{
	//		{
	//			SnapshotId: snapShot.SnapshotId,
	//		},
	//	},
	//}
	cvi := &ec2.CreateVolumeInput{
		//ToDO Dynamically retrieve from forensic_subnet-id
		AvailabilityZone: aws.String("ap-northeast-1a"),
		SnapshotId: snapShot.SnapshotId,
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("volume"),
				Tags: []*ec2.Tag{
					{
						Key: aws.String("Name"), Value:aws.String("forensic-ebs-volume"),
					},
				},
			},
		},
	}
	cvo, err := svc.CreateVolume(cvi)
	//io, err := svc.ImportImage(ii) // CreateImage
	if err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "create an EBS volume").Msg("failed")
	}

	// Check Status
	//var importStatus string
	//for importStatus != "completed"  {
	//	do, err := svc.DescribeVolumeta(&ec2.DescribeImportImageTasksInput{
	//		ImportTaskIds: []*string{io.ImportTaskId},
	//	})
	//	if err != nil {
	//		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "create an AMI").Msg("failed")
	//	}
	//	importStatus = *do.ImportImageTasks[0].Status
	//}
	log.Info().Str("duration", returnDuration()).Str("status", "create an EBS volume").Msg("succeeded")

	// Create Isolated Security Group
	//log.Info().Str("duration", returnDuration()).Str("status", "create a SG").Msg("succeeded")
	//csgi := &ec2.CreateSecurityGroupInput{
	//	VpcId: aws.String(forensicVpcId),
	//	GroupName: aws.String("forensic-isolation-sg"),
	//}
	//csgo, err := svc.CreateSecurityGroup(csgi)
	//if err != nil {
	//	log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "create a SG").Msg("failed")
	//}
	//if _, err := svc.AuthorizeSecurityGroupEgress(&ec2.AuthorizeSecurityGroupEgressInput{
	//	GroupId: csgo.GroupId,
	//	IpPermissions:[]*ec2.IpPermission{},
	//}); err != nil {
	//	log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "create a SG").Msg("failed")
	//}
	//log.Info().Str("duration", returnDuration()).Str("status", "create a SG").Msg("completed")

	// Run Instance in Forensic VPC
	log.Info().Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("started")
	ro := &ec2.RunInstancesInput{
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key: aws.String("Name"), Value:aws.String("forensic-instance"),
					},
				},
			},
		},
		MaxCount: aws.Int64(1),
		MinCount: aws.Int64(1),
		SecurityGroupIds: []*string{aws.String(forensicSgId)},
		SubnetId: aws.String(forensicSubnetId),
		InstanceType: aws.String("t2.micro"), //TODO put into config // maybe t2.large is better according to https://www.sans.org/reading-room/whitepapers/cloud/digital-forensic-analysis-amazon-linux-ec2-instances-38235?
		ImageId: aws.String("ami-e99f4896"), //ToDo Add User Data or make it the latest
	}

	re, err := svc.RunInstances(ro)
	if err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("failed")
	}
	log.Info().Str("duration", returnDuration()).Str("status", "Starting up a forensic instance").Msg("succeeded")

	log.Info().Str("duration", returnDuration()).Str("status", "Attaching Volume").Msg("started")
	if _, err := svc.AttachVolume(&ec2.AttachVolumeInput{
		InstanceId: re.Instances[0].InstanceId,
		VolumeId: cvo.VolumeId,
		Device: aws.String("/dev/sdf"), //TODO Change device name based on already running instance.
	}); err != nil {
		log.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "Attaching Volume").Msg("failed")
	}
	log.Info().Str("duration", returnDuration()).Str("status", "Attaching Volume").Msg("succeeded")

	return fmt.Sprintf("Created Snapshot %s!", snapShot.SnapshotId), nil
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

