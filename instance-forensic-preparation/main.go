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
var logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})

var slackURL, forensicVPC string

func init() {
	slackURL = os.Getenv("SLACK_URL")
	forensicVPC = os.Getenv("FORENSIC_VPC")
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	if slackURL == "" || forensicVPC == "" {
		logger.Fatal().Msg("you must set Env Var `SLACK_URL` and `FORENSIC_URL`")
	}
	lambda.Start(HandleRequest)
}
//func HandleRequest(request CloudWatchEventForGuardDuty) (string, error) {
func HandleRequest(instanceId string) (string, error) {
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
		logger.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "stopping instance").Msg("failed")
	} else {
		if err := svc.WaitUntilInstanceStoppedWithContext(ctx, describeInstancesInput); err != nil {
			logger.Fatal().Err(err).Str("duration", returnDuration()).Str("status", "stopping instance").Msg("failed")
		}
	}
	log.Info().Str("duration", returnDuration()).Str("status", "stop instance").Msg("succeeded")

	// Describe Instance
	log.Info().Str("duration", returnDuration()).Str("status", "describe instance").Msg("started")
	out, err := svc.DescribeInstanceAttributeWithContext(ctx, describeEc2AttributeInput)
	if err != nil {
		log.Error().Err(err).Str("duration", returnDuration()).Str("status", "describe instance").Msg("failed")
		return "", err
	}
	log.Info().Str("duration", returnDuration()).Str("status", "describe instance").Msg("succeeded")

	/*
	//  Create a Snapshot
	 */
	// assuming there is only one volume
	createSnapshotInput.VolumeId = out.BlockDeviceMappings[0].Ebs.VolumeId
	createSnapshotInput.TagSpecifications = []*ec2.TagSpecification{
		{
			ResourceType: aws.String("instance"),
			Tags: []*ec2.Tag{{Key: aws.String("hoge"), Value:aws.String("fuga")}},
		},
	}
	log.Info().Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("started")
	snapShot, err := svc.CreateSnapshotWithContext(ctx, createSnapshotInput)
	if err != nil {
		log.Error().Err(err).Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("failed")
		return "", err
	}
	log.Info().Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("succeeded")

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

