provider "aws" {
  alias  = "master-tokyo"
  region = "${var.aws-regions["tokyo"]}"
}

provider "aws" {
  alias  = "master-virginia"
  region = "${var.aws-regions["virginia"]}"
}

provider "aws" {
  alias  = "master-ohio"
  region = "${var.aws-regions["ohio"]}"
}

provider "aws" {
  alias  = "master-california"
  region = "${var.aws-regions["california"]}"
}

provider "aws" {
  alias  = "master-oregon"
  region = "${var.aws-regions["oregon"]}"
}

provider "aws" {
  alias  = "master-seou"
  region = "${var.aws-regions["seoul"]}"
}

provider "aws" {
  alias  = "master-osaka"
  region = "${var.aws-regions["osaka"]}"
}

provider "aws" {
  alias  = "master-mumbai"
  region = "${var.aws-regions["mumbai"]}"
}

provider "aws" {
  alias  = "master-singapore"
  region = "${var.aws-regions["singapore"]}"
}

provider "aws" {
  alias  = "master-sydney"
  region = "${var.aws-regions["sydney"]}"
}

provider "aws" {
  alias  = "master-canada"
  region = "${var.aws-regions["canada"]}"
}

provider "aws" {
  alias  = "master-beijing"
  region = "${var.aws-regions["beijing"]}"
}

provider "aws" {
  alias  = "master-ningxia"
  region = "${var.aws-regions["ningxia"]}"
}

provider "aws" {
  alias  = "master-frankfurt"
  region = "${var.aws-regions["frankfurt"]}"
}

provider "aws" {
  alias  = "master-ireland"
  region = "${var.aws-regions["ireland"]}"
}

provider "aws" {
  alias  = "master-london"
  region = "${var.aws-regions["london"]}"
}

provider "aws" {
  alias  = "master-paris"
  region = "${var.aws-regions["paris"]}"
}

provider "aws" {
  alias  = "master-sao-paulo"
  region = "${var.aws-regions["sao-paulo"]}"
}

provider "aws" {
  alias   = "member001-tokyo"
  region  = "ap-northeast-1"
  profile = "sub"
}

// Tokyo Region GD
resource "aws_guardduty_detector" "master-tokyo" {
  provider = "aws.master-tokyo"
  enable   = true
}

resource "aws_guardduty_detector" "member001-tokyo" {
  provider = "aws.member001-tokyo"
  enable   = true
}

// Virginia Region GD
resource "aws_guardduty_detector" "master-virginia" {
  provider = "aws.master-virginia"
  enable   = true
}

//
//resource "aws_guardduty_member" "member" {
//  account_id = "${aws_guardduty_detector.member.account_id}"
//  detector_id = "${aws_guardduty_detector.master.id}"terraform import aws_guardduty_detector.MyDetector
//  email = "kengoscal+001@gmail.com"
//  invite = true
//  invitation_message = "hogefugater"
//}

