

# Pre-requisite
* Incoming Webhook URL for Slack
* VPC for forensic environment
* Private subnet in forensic VPC
* Security Group with no egress traffic allowerd in forensic VPC
* Put them in Lambda's environment variable
  * SLACK_URL
  * FORENSIC_VPC_ID
  * FORENSIC_SUBNET_ID
  * FORENSIC_SG_ID


# Build
* `OS=linux GOARCH=amd64 go build -o main && zip main.zip ./main`


