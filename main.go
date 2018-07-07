package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"github.com/nlopes/slack"
	"encoding/json"
	"bytes"
	"os"
)

var ErrNameNotProvided = errors.New("no name was provided in the HTTP body")
var ErrSlackPostingFailed = errors.New("posting Slack failed")
var slackURL string
var contentType = "application/json"

func init() {
	slackURL = os.Getenv("SLACK_URL")
}

//https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/
func main() {
	lambda.Start(HandleRequest)
	fmt.Println("test check")
}

func HandleRequest(request GuardDutyRequest) (string, error) {
	if slackURL == "" {
		return "", errors.New("you must set Env Var `SLACK_URL`")
	}

	if err := postOnSlack(request); err != nil {
		return "", err
	}
	return fmt.Sprintf("Hello %s!", "hogefuga"), nil
}

func postOnSlack(request GuardDutyRequest) error {
	severity, err := mapSeverityToLevel(request)
	if err != nil {
		return err
	}

	pretext := "'" + request.Detail.Type + "' type found."
	if severity.Announce {
		pretext = "<!here> " + pretext
	}


	attachment := slack.Attachment{
		Color: severity.Color,
		Pretext: pretext,
		Title: request.Detail.Title,
		Fields: []slack.AttachmentField{
			{
				Title: "Severity",
				Value: severity.Level,
				Short: true,
			},
			{
				Title: "Account",
				Value: severity.AccountAlias,
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

type Severity struct {
	AccountAlias string
	Level string
	Color string
	Announce bool
}

func mapSeverityToLevel(request GuardDutyRequest) (*Severity, error) {
	severity := request.Detail.Severity
	s := &Severity{}

	//if alias, ok := accountMap[request.Detail.AccountID]; ok {
	//	s.AccountAlias = alias
	//} else {
	//	s.AccountAlias = request.Account + ": Alias Not Found"
	//}
	s.AccountAlias = request.Detail.AccountID

	if request.Detail_type == "Recon:EC2/PortProbeUnprotectedPort" {
		gd := &GuardDutyRequest{Detail:GuardDutyRequestDetail{Service: PortProbeAction{}}}

		fmt.Println(gd.Detail.Service.(PortProbeAction).Count)
	}

	if 0 <= severity && severity < 4 {
		s.Level = "Low"
		s.Color = "#707070"
		return s, nil
	} else if 4 <= severity && severity < 7 {
		s.Level = "Medium"
		s.Color = "warning"
		return s, nil
	} else if severity < 10 {
		s.Level = "High"
		s.Color = "danger"
		s.Announce = true
		return s, nil
	}

	return nil, errors.New("Severity was not in right range: 0~10.0")
}

type ProbeEvents struct {
	EventLastSeen  string `json:"eventLastSeen"`
	Count          int    `json:"count"`
}

type GuardDutyRequest struct {
	Account string `json:"account"`
	Detail  GuardDutyRequestDetail `json:"detail"`
	Detail_type string        `json:"detail-type"`
	ID          string        `json:"id"`
	Region      string        `json:"region"`
	Resources   []interface{} `json:"resources"`
	Source      string        `json:"source"`
	Time        string        `json:"time"`
	Version     string        `json:"version"`
}

type GuardDutyRequestDetail struct {
	// Parameters that are important
	Severity    float64     `json:"severity"`
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
}

type PortProbeAction struct {
	Action      struct {
		ActionType      string `json:"actionType"`
		PortProbeAction struct {
			PortProbeDetails []struct {
				LocalPortDetails struct {
					Port     int    `json:"port"`
					PortName string `json:"portName"`
				} `json:"localPortDetails"`
				RemoteIPDetails struct {
					Country struct {
						CountryName string `json:"countryName"`
					} `json:"country"`
					City struct {
						CityName string `json:"cityName"`
					} `json:"city"`
					GeoLocation struct {
						Lon float64 `json:"lon"`
						Lat float64 `json:"lat"`
					} `json:"geoLocation"`
					Organization struct {
						AsnOrg string `json:"asnOrg"`
						Org    string `json:"org"`
						Isp    string `json:"isp"`
						Asn    string `json:"asn"`
					} `json:"organization"`
					IPAddressV4 string `json:"ipAddressV4"`
				} `json:"remoteIpDetails"`
			} `json:"portProbeDetails"`
			Blocked bool `json:"blocked"`
		} `json:"portProbeAction"`
	} `json:"action"`
	AdditionalInfo struct {
		ThreatListName  string `json:"threatListName"`
		ThreatName         int    `json:"threatName"`
	} `json:"additionalInfo"`
	Archived       bool   `json:"archived"`
	Count          int    `json:"count"`
	DetectorID     string `json:"detectorId"`
	EventFirstSeen string `json:"eventFirstSeen"`
	EventLastSeen  string `json:"eventLastSeen"`
	ResourceRole   string `json:"resourceRole"`
	ServiceName    string `json:"serviceName"`
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
