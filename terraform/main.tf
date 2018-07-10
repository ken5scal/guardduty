provider "aws" {
  region = "ap-northeast-1"
}

provider "aws" {
  alias   = "member"
  region  = "ap-northeast-1"
  profile = "sub"
}

//resource "aws_guardduty_detector" "master" {
//  enable = true
//}
//
//resource "aws_guardduty_detector" "member" {
//  provider = "aws.member"
//  enable = true
//}
//
//resource "aws_guardduty_member" "member" {
//  account_id = "${aws_guardduty_detector.member.account_id}"
//  detector_id = "${aws_guardduty_detector.master.id}"
//  email = "kengoscal+001@gmail.com"
//  invite = true
//  invitation_message = "hogefugater"
//}

resource "aws_instance" "web" {
  ami           = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"

  tags {
    Name = "HelloWorld"
  }
}

resource "aws_iam_policy" "example" {
  name   = "example_policy"
  path   = "/"
  policy = "${data.aws_iam_policy_document.example.json}"
}

data "aws_iam_policy_document" "example" {
  statement {
    actions   = ["s3:ListAllMyBuckets"]
    resources = ["arn:aws:s3:::*"]
  }
}
