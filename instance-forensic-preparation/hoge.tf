variable "snapshot_id" {}

variable "is_incident" {
  default = 0
}

variable "clean_room_cidr" {
  default = "172.32.0.0/24"
}

resource "aws_vpc" "clean_room_vpc" {
  count = "${var.is_incident}"
  cidr_block = "${var.clean_room_cidr}"
  tags {
    Name = "CleanRoomVPC"
  }
}

resource "aws_security_group" "clean_room_sg" {
  name        = "forensic_sg"
  description = "Allow no rule. Just your IP or bastion IP"
  vpc_id = "${aws_vpc.clean_room_vpc.id}"
  // There will be no rule
  // ingress {}
  // egress {}

  tags {
    Name = "CleanRoomSG"
  }
}

resource "aws_ami" "investigated_ami" {
  name = "investigated_ami"
  root_snapshot_id = "${var.snapshot_id}"
}

resource "aws_instance" "investifate" {
  ami = "${aws_ami.investigated_ami}"
  instance_type = "t2.small"
  vpc_security_group_ids = ["${aws_security_group.clean_room_sg.id}"]
}