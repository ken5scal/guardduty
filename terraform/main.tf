provider "aws" {
  region = "ap-northeast-1"
}

provider "aws" {
  alias = "member"
  region = "ap-northeast-1"
}

resource "aws_guardduty_detector" "main" {
  enable = true
}

resource "aws_guardduty_detector" "member" {
  provider = "aws.member"
  enable = true
}

resource "aws_guardduty_member" "member" {
  account_id = "${aws_guardduty_detector.member.account_id}"
  detector_id = "${aws_guardduty_detector.main.id}"
  email = "kengoscal+001@gmail.com"
  invite = true
  invitation_message = "hogefuga"
}
