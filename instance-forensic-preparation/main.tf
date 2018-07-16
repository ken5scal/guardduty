provider "aws" {
  region = "ap-northeast-1"
}

variable "account_id" {}
variable "slack_url" {}

variable "region" {
  default = "ap-northeast-1"
}

variable "forensic_vpc_cidr" {
  default = "172.32.0.0/28"
}

variable "trusted_cidr" {
  default = "119.106.15.81/32"
}

resource "aws_iam_role" "forensic_lambda_role" {
  name = "forensic_role_for_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "forensic_lambda_policy" {
  name        = "forensic-policy"
  description = "Policy required for forensic"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAllEC2NoneDestructivePolicies",
      "Effect": "Allow",
      "Action": [
                "ec2:DescribeInstances",
                "ec2:DescribeVolumesModifications",
                "ec2:GetHostReservationPurchasePreview",
                "ec2:DescribeSnapshots",
                "ec2:DescribePlacementGroups",
                "ec2:GetConsoleScreenshot",
                "ec2:DescribeHostReservationOfferings",
                "ec2:DescribeInternetGateways",
                "ec2:GetLaunchTemplateData",
                "ec2:DescribeVolumeStatus",
                "ec2:DescribeScheduledInstanceAvailability",
                "ec2:DescribeSpotDatafeedSubscription",
                "ec2:DescribeVolumes",
                "ec2:DescribeFpgaImageAttribute",
                "ec2:DescribeExportTasks",
                "ec2:DescribeAccountAttributes",
                "ec2:DescribeNetworkInterfacePermissions",
                "ec2:DescribeReservedInstances",
                "ec2:DescribeKeyPairs",
                "ec2:DescribeNetworkAcls",
                "ec2:DescribeRouteTables",
                "ec2:DescribeReservedInstancesListings",
                "ec2:DescribeEgressOnlyInternetGateways",
                "ec2:DescribeSpotFleetRequestHistory",
                "ec2:DescribeLaunchTemplates",
                "ec2:DescribeVpcClassicLinkDnsSupport",
                "ec2:DescribeVpnConnections",
                "ec2:DescribeSnapshotAttribute",
                "ec2:DescribeVpcPeeringConnections",
                "ec2:DescribeReservedInstancesOfferings",
                "ec2:DescribeIdFormat",
                "ec2:DescribeVpcEndpointServiceConfigurations",
                "ec2:DescribePrefixLists",
                "ec2:GetReservedInstancesExchangeQuote",
                "ec2:DescribeVolumeAttribute",
                "ec2:DescribeInstanceCreditSpecifications",
                "ec2:DescribeVpcClassicLink",
                "ec2:DescribeImportSnapshotTasks",
                "ec2:DescribeVpcEndpointServicePermissions",
                "ec2:GetPasswordData",
                "ec2:DescribeScheduledInstances",
                "ec2:DescribeImageAttribute",
                "ec2:DescribeVpcEndpoints",
                "ec2:DescribeReservedInstancesModifications",
                "ec2:DescribeElasticGpus",
                "ec2:DescribeSubnets",
                "ec2:DescribeVpnGateways",
                "ec2:DescribeMovingAddresses",
                "ec2:DescribeAddresses",
                "ec2:DescribeInstanceAttribute",
                "ec2:DescribeRegions",
                "ec2:DescribeFlowLogs",
                "ec2:DescribeDhcpOptions",
                "ec2:DescribeVpcEndpointServices",
                "ec2:DescribeSpotInstanceRequests",
                "ec2:DescribeVpcAttribute",
                "ec2:GetConsoleOutput",
                "ec2:DescribeSpotPriceHistory",
                "ec2:DescribeNetworkInterfaces",
                "ec2:DescribeAvailabilityZones",
                "ec2:DescribeNetworkInterfaceAttribute",
                "ec2:DescribeVpcEndpointConnections",
                "ec2:DescribeInstanceStatus",
                "ec2:DescribeHostReservations",
                "ec2:DescribeIamInstanceProfileAssociations",
                "ec2:DescribeTags",
                "ec2:DescribeLaunchTemplateVersions",
                "ec2:DescribeBundleTasks",
                "ec2:DescribeIdentityIdFormat",
                "ec2:DescribeImportImageTasks",
                "ec2:DescribeClassicLinkInstances",
                "ec2:DescribeNatGateways",
                "ec2:DescribeCustomerGateways",
                "ec2:DescribeVpcEndpointConnectionNotifications",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeSpotFleetRequests",
                "ec2:DescribeHosts",
                "ec2:DescribeImages",
                "ec2:DescribeFpgaImages",
                "ec2:DescribeSpotFleetInstances",
                "ec2:DescribeSecurityGroupReferences",
                "ec2:DescribeVpcs",
                "ec2:DescribeConversionTasks",
                "ec2:DescribeStaleSecurityGroups"
      ],
      "Resource": "*"
    },
    {
        "Sid": "AllowMinPrivilegeForWritingEC2",
        "Effect": "Allow",
        "Action": [
            "ec2:AttachVolume",
            "ec2:CreateTags",
            "ec2:CreateSnapshot",
            "ec2:RunInstances",
            "ec2:ImportImage",
            "ec2:StopInstances",
            "ec2:CreateVolume"
        ],
        "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "basic_lambda_policy" {
  name        = "basic-lambda-policy"
  description = "Policy required for basic lambda functionality"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "VisualEditor1",
      "Effect": "Allow",
      "Action": [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:${var.region}:${var.account_id}:log-group:${aws_cloudwatch_log_group.forensic_lambda_log_group.name}:*"
    },
    {
      "Sid": "VisualEditor2",
      "Effect": "Allow",
      "Action": "logs:CreateLogGroup",
      "Resource": "arn:aws:logs:${var.region}:${var.account_id}:*"
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "forensic_lambda_log_group" {
  name = "/aws/lambda/forensic-lambda"
}

resource "aws_iam_role_policy_attachment" "forensic_lambda" {
  role       = "${aws_iam_role.forensic_lambda_role.name}"
  policy_arn = "${aws_iam_policy.forensic_lambda_policy.arn}"
}

resource "aws_iam_role_policy_attachment" "basic_lambda" {
  role       = "${aws_iam_role.forensic_lambda_role.name}"
  policy_arn = "${aws_iam_policy.basic_lambda_policy.arn}"
}

resource "aws_lambda_function" "forensic_lambda" {
  function_name = "forensic-lambda"
  filename = "main.zip"
  handler       = "main"
  role          = "${aws_iam_role.forensic_lambda_role.arn}"
  runtime       = "go1.x"
  timeout = 300

  environment {
    variables {
      FORENSIC_SG_ID = "${aws_security_group.forensic_isolated_sg.id}"
      FORENSIC_SUBNET_ID = "${aws_subnet.forensic_private_subnet.id}"
      FORENSIC_VPC_ID = "${aws_vpc.forensic_vpc.id}"
      SLACK_URL = "${var.slack_url}"
    }
  }
}

resource "aws_cloudwatch_event_rule" "send_all_guardduty_events_rule" {
  name = "RuleToSendAllGuardDutyEvents"
  event_pattern = <<EOF
{
  "detail-type": [
    "GuardDuty Finding"
  ],
  "source": [
    "aws.guardduty"
  ]
}
EOF
}

resource "aws_cloudwatch_event_target" "guardduty-lambda" {
  rule = "${aws_cloudwatch_event_rule.send_all_guardduty_events_rule.name}"
  arn = "${aws_lambda_function.forensic_lambda.arn}"
}

resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id   = "AllowExecutionFromCloudWatch"
  action         = "lambda:InvokeFunction"
  function_name  = "${aws_lambda_function.forensic_lambda.function_name}"
  principal      = "events.amazonaws.com"
  source_arn = "${aws_cloudwatch_event_rule.send_all_guardduty_events_rule.arn}"
}

resource "aws_vpc" "forensic_vpc" {
  cidr_block = "${var.forensic_vpc_cidr}"

  tags {
    Name = "Forensic-Vpc"
  }
}

resource "aws_subnet" "forensic_private_subnet" {
  cidr_block = "${var.forensic_vpc_cidr}"
  vpc_id     = "${aws_vpc.forensic_vpc.id}"
  availability_zone = "ap-northeast-1a"

  tags {
    Name = "Forensic-Private-Subnet"
  }
}

resource "aws_security_group" "forensic_isolated_sg" {
  name        = "forensic_isolated_sg"
  description = "Allow no rule. Just your IP or bastion IP"
  vpc_id      = "${aws_vpc.forensic_vpc.id}"

  egress = []

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${var.trusted_cidr}"]
  }

  tags {
    Name = "Forensic-Isolated-Sg"
  }
}

