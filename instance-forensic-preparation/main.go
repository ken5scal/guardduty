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
	DryRun:      aws.Bool(true),
	Force:       aws.Bool(true),
	InstanceIds: []*string{},
}
var describeInstancesInput = &ec2.DescribeInstancesInput{
	DryRun:      aws.Bool(true),
	InstanceIds: []*string{},
}
var copySnapshotInput = &ec2.CopySnapshotInput{
	Description: aws.String("Snapshot taken for forensic purpose"),
	DryRun:      aws.Bool(true),
	//Encrypted: aws.Bool(true), // for now
	SourceRegion: aws.String("ap-northeast-1"),
}
var describeEc2AttributeInput = &ec2.DescribeInstanceAttributeInput{
	DryRun:    aws.Bool(true),
	Attribute: aws.String("blockDeviceMapping"),
}
var logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})

func main() {
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
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
	ctx, cancelFn := context.WithTimeout(context.Background(), timeout)
	defer cancelFn()

	log.Info().Str("duration", returnDuration()).Str("status", "stop instance").Msg("started")
	if _, err := svc.StopInstancesWithContext(ctx, stopInstancesInput); err != nil {
		log.Error().Err(err).Str("duration", returnDuration()).Str("status", "stopping instance").Msg("failed")
		return "", err
	}

	if err := svc.WaitUntilInstanceStoppedWithContext(ctx, describeInstancesInput); err != nil {
		log.Error().Err(err).Str("duration", returnDuration()).Str("status", "stopping instance").Msg("failed")
		return "", err
	}
	log.Info().Str("duration", returnDuration()).Str("status", "stop instance").Msg("succeeded")

	log.Info().Str("duration", returnDuration()).Str("status", "describe instance").Msg("started")
	out, err := svc.DescribeInstanceAttributeWithContext(ctx, describeEc2AttributeInput)
	if err != nil {
		log.Error().Err(err).Str("duration", returnDuration()).Str("status", "describe instance").Msg("failed")
		return "", err
	}
	log.Info().Str("duration", returnDuration()).Str("status", "describe instance").Msg("succeeded")

	// assuming there is only one volume
	copySnapshotInput.SourceSnapshotId = out.BlockDeviceMappings[0].Ebs.VolumeId
	log.Info().Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("started")
	if _, err := svc.CopySnapshotWithContext(ctx, copySnapshotInput); err != nil {
		log.Error().Err(err).Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("failed")
		return "", err
	}
	log.Info().Str("duration", returnDuration()).Str("status", "taking snapshot").Msg("succeeded")

	return fmt.Sprintf("Copied Snapshot %s!", copySnapshotInput.SourceSnapshotId), nil
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