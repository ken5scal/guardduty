provider "aws" {
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
  alias  = "master-seoul"
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

resource "aws_guardduty_detector" "master-tokyo" {
  enable   = true
}

resource "aws_guardduty_detector" "master-ohio" {
  provider = "aws.master-ohio"
  enable   = true
}

resource "aws_guardduty_detector" "master-virginia" {
  provider = "aws.master-virginia"
  enable   = true
}

resource "aws_guardduty_detector" "master-california" {
  provider = "aws.master-california"
  enable   = true
}

resource "aws_guardduty_detector" "master-oregon" {
  provider = "aws.master-oregon"
  enable   = true
}

resource "aws_guardduty_detector" "master-seoul" {
  provider = "aws.master-seoul"
  enable   = true
}

//Not yet available
//resource "aws_guardduty_detector" "master-osaka" {
//  provider = "aws.master-osaka"
//  enable   = true
//}

resource "aws_guardduty_detector" "master-mumbai" {
  provider = "aws.master-mumbai"
  enable   = true
}

resource "aws_guardduty_detector" "master-singapore" {
  provider = "aws.master-singapore"
  enable   = true
}

resource "aws_guardduty_detector" "master-sydney" {
  provider = "aws.master-sydney"
  enable   = true
}

resource "aws_guardduty_detector" "master-canada" {
  provider = "aws.master-canada"
  enable   = true
}

//Not yet available
//resource "aws_guardduty_detector" "master-beijing" {
//  provider = "aws.master-beijing"
//  enable   = true
//}
//
//Not yet available
//resource "aws_guardduty_detector" "master-ningxia" {
//  provider = "aws.master-ningxia"
//  enable   = true
//}

resource "aws_guardduty_detector" "master-frankfurt" {
  provider = "aws.master-frankfurt"
  enable   = true
}

resource "aws_guardduty_detector" "master-ireland" {
  provider = "aws.master-ireland"
  enable   = true
}

resource "aws_guardduty_detector" "master-london" {
  provider = "aws.master-london"
  enable   = true
}

resource "aws_guardduty_detector" "master-paris" {
  provider = "aws.master-paris"
  enable   = true
}

resource "aws_guardduty_detector" "master-sao-paulo" {
  provider = "aws.master-sao-paulo"
  enable   = true
}

resource "aws_guardduty_detector" "member001-tokyo" {
  provider = "aws.member001-tokyo"
  enable   = true
}

resource "aws_guardduty_member" "member001-tokyo" {
  account_id = "${aws_guardduty_detector.member001-tokyo.account_id}"
  detector_id = "${aws_guardduty_detector.master-tokyo.id}"
  email = "kengoscal+001@gmail.com"
  invite = true
  invitation_message = "hogefugater"
}