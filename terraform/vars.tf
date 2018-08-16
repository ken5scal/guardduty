variable "aws-regions" {
  type = "map"

  default = {
    ohio       = "us-east-2"
    virginia   = "us-east-1"
    california = "us-west-1"
    oregon     = "us-west-2"
    tokyo      = "ap-northeast-1"
    seoul      = "ap-northeast-2"
    osaka      = "ap-northeast-3"
    mumbai     = "ap-south-1"
    singapore  = "ap-southeast-1"
    sydney     = "ap-southeast-2"
    canada     = "ca-central-1"
    beijing    = "cn-north-1"
    ningxia    = "cn-northwest-1"
    frankfurt  = "eu-central-1"
    ireland    = "eu-west-1"
    london     = "eu-west-2"
    paris      = "eu-west-3"
    sao-paulo  = "sa-east-1"
  }
}
