package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"net/http"
	"github.com/nlopes/slack"
	"encoding/json"
	"bytes"
)

var ErrNameNotProvided = errors.New("no name was provided in the HTTP body")
var ErrSlackPostingFailed = errors.New("posting Slack failed")
var slackURL = "https://hooks.slack.com/services/"
var contentType = "application/json"

//https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/
func main() {
	lambda.Start(HandleRequest)
}

// Handler is your Lambda function handler
// It uses Amazon API Gateway request/responses provided by the aws-lambda-go/events package,
// However you could use other event sources (S3, Kinesis etc), or JSON-decoded primitive types such as 'string'.
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Processing Lambda request %s\n", request.RequestContext.RequestID)

	// If no name is provided in the HTTP request body, throw an error
	if len(request.Body) < 1 {
		return events.APIGatewayProxyResponse{}, ErrNameNotProvided
	}

	return events.APIGatewayProxyResponse{
		Body:       "Hi " + request.Body,
		StatusCode: 200,
	}, nil

}

func HandleRequest(request GuardDutyRequest) (string, error) {
	if err := postOnSlack(request); err != nil {
		return "", err
	}
	return fmt.Sprintf("Hello %s!", "hogefuga"), nil
}

func postOnSlack(request GuardDutyRequest) error {
	attachment := slack.Attachment{
		Color: "danger", //warning, good, pretext
		Pretext: "'" + request.Detail.Type + "' type found.",
		Title: request.Detail.Title,
		Fields: []slack.AttachmentField{
			{
				Title: "Severity",
				Value: "High",  // TODO Map to High, Medium, Low from request.Detail.Severity
				Short: true,
			},
			{
				Title: "Account",
				Value: request.Account,  // TODO Map to account names
				Short: true,
			},
			{
				Title: "Description",
				Value: request.Detail.Description,
				Short: false,
			},
		},
	}

	payload := slack.PostMessageParameters{
		Attachments: []slack.Attachment{attachment},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return errors.New("failed to encode payload")
	}

	if _, err := http.Post(slackURL, contentType, bytes.NewReader(body)); err != nil {
		return errors.New(fmt.Sprintf("posting Slack failed: %v", err))
	}

	return nil
}


type GuardDutyRequest struct {
	Account string `json:"account"`
	Detail  struct {
		// Parameters that are important
		Severity    int         `json:"severity"`
		Title       string      `json:"title"`
		Type        string      `json:"type"`
		UpdatedAt   string      `json:"updatedAt"`
		AccountID   string      `json:"accountId"`
		Description string      `json:"description"`
		Resource    interface{} `json:"resource"`
		Service     interface{} `json:"service"`
		//Resource  InstanceResource `json:"resource"`
		//Service       NetworkConnectionAction `json:"service"`

		// Parameters that are not so important
		Region        string `json:"region"`
		Arn           string `json:"arn"`
		CreatedAt     string `json:"createdAt"`
		ID            string `json:"id"`
		Partition     string `json:"partition"`
		SchemaVersion string `json:"schemaVersion"`
	} `json:"detail"`
	Detail_type string        `json:"detail-type"`
	ID          string        `json:"id"`
	Region      string        `json:"region"`
	Resources   []interface{} `json:"resources"`
	Source      string        `json:"source"`
	Time        string        `json:"time"`
	Version     string        `json:"version"`
}

type NetworkConnectionAction struct {
	Action struct {
		ActionType              string `json:"actionType"`
		NetworkConnectionAction struct {
			Blocked             bool   `json:"blocked"`
			ConnectionDirection string `json:"connectionDirection"`
			LocalPortDetails    struct {
				Port     int    `json:"port"`
				PortName string `json:"portName"`
			} `json:"localPortDetails"`
			Protocol        string `json:"protocol"`
			RemoteIPDetails struct {
				City struct {
					CityName string `json:"cityName"`
				} `json:"city"`
				Country struct {
					CountryName string `json:"countryName"`
				} `json:"country"`
				GeoLocation struct {
					Lat int `json:"lat"`
					Lon int `json:"lon"`
				} `json:"geoLocation"`
				IPAddressV4 string `json:"ipAddressV4"`

				Organization struct {
					Asn int    `json:"asn"`
					Isp string `json:"isp"`
					Org string `json:"org"`
				} `json:"organization"`
			} `json:"remoteIpDetails"`
			RemotePortDetails struct {
				Port     int    `json:"port"`
				PortName string `json:"portName"`
			} `json:"remotePortDetails"`
		} `json:"networkConnectionAction"`
	} `json:"action"`
	AdditionalInfo struct {
		ThreatListName  string `json:"threatListName"`
		Unusual         int    `json:"unusual"`
		UnusualProtocol string `json:"unusualProtocol"`
	} `json:"additionalInfo"`
	Archived       bool   `json:"archived"`
	Count          int    `json:"count"`
	DetectorID     string `json:"detectorId"`
	EventFirstSeen string `json:"eventFirstSeen"`
	EventLastSeen  string `json:"eventLastSeen"`
	ResourceRole   string `json:"resourceRole"`
	ServiceName    string `json:"serviceName"`
}

type InstanceResource struct {
	ResourceType    string `json:"resourceType"`
	InstanceDetails struct {
		AvailabilityZone  string `json:"availabilityZone"`
		ImageDescription  string `json:"imageDescription"`
		ImageID           string `json:"imageId"`
		InstanceID        string `json:"instanceId"`
		InstanceState     string `json:"instanceState"`
		InstanceType      string `json:"instanceType"`
		LaunchTime        int    `json:"launchTime"`
		NetworkInterfaces []struct {
			Ipv6Addresses      []interface{} `json:"ipv6Addresses"`
			PrivateDNSName     string        `json:"privateDnsName"`
			PrivateIPAddress   string        `json:"privateIpAddress"`
			PrivateIPAddresses []struct {
				PrivateDNSName   string `json:"privateDnsName"`
				PrivateIPAddress string `json:"privateIpAddress"`
			} `json:"privateIpAddresses"`
			PublicDNSName  string `json:"publicDnsName"`
			PublicIP       string `json:"publicIp"`
			SecurityGroups []struct {
				GroupID   string `json:"groupId"`
				GroupName string `json:"groupName"`
			} `json:"securityGroups"`
			SubnetID string `json:"subnetId"`
			VpcID    string `json:"vpcId"`
		} `json:"networkInterfaces"`
		ProductCodes []interface{} `json:"productCodes"`
		Tags         []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"tags"`
	} `json:"instanceDetails"`
}
