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

func main() {
	lambda.Start(HandleRequest)
}

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

	if _, err := svc.StopInstancesWithContext(ctx, stopInstancesInput); err != nil {
		return "", err
	}

	if err := svc.WaitUntilInstanceStoppedWithContext(ctx, describeInstancesInput); err != nil {
		return "", err
	}

	out, err := svc.DescribeInstanceAttributeWithContext(ctx, describeEc2AttributeInput)
	if err != nil {
		return "", err
	}

	// assuming there is only one volume
	copySnapshotInput.SourceSnapshotId = out.BlockDeviceMappings[0].Ebs.VolumeId
	if _, err := svc.CopySnapshotWithContext(ctx, copySnapshotInput); err != nil {
		return "", err
	}

	return fmt.Sprintf("Copied Snapshot %s!", copySnapshotInput.SourceSnapshotId), nil
}
