package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nlopes/slack"
	"net/http"
	"os"
	"github.com/rs/zerolog/log"
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
}

func HandleRequest(request CloudWatchEventForGuardDuty) {
	if slackURL == "" {
		log.Fatal().Msg("you must set Env Var `SLACK_URL`")
	}

	if err := postOnSlack(request); err != nil {
		log.Fatal().Err(err).Msg("failed.")
	}
}

func postOnSlack(request CloudWatchEventForGuardDuty) error {
	severity, err := mapSeverityToLevel(request)
	if err != nil {
		return err
	}

	pretext := "'" + request.Detail.Type + "' type found."
	if severity.Announce {
		pretext = "<!here> " + pretext
	}

	attachment := slack.Attachment{
		Color:   severity.Color,
		Pretext: pretext,
		Title:   request.Detail.Title,
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
	Level        string
	Color        string
	Announce     bool
}

func mapSeverityToLevel(request CloudWatchEventForGuardDuty) (*Severity, error) {
	severity := request.Detail.Severity
	s := &Severity{}
	s.AccountAlias = request.Detail.AccountID

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

// CloudWatchEventForGuardDuty: https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_findings_cloudwatch.html
type CloudWatchEventForGuardDuty struct {
	Account     string           `json:"account"`
	Detail      GuardDutyFinding `json:"detail"`
	Detail_type string           `json:"detail-type"`
	ID          string           `json:"id"`
	Region      string           `json:"region"`
	Resources   []interface{}    `json:"resources"`
	Source      string           `json:"source"`
	Time        string           `json:"time"`
	Version     string           `json:"version"`
}

// https://docs.aws.amazon.com/guardduty/latest/ug/get-findings.html#get-findings-response-syntax
type GuardDutyFinding struct {
	// Parameters that are not so important
	SchemaVersion string `json:"schemaVersion"`
	AccountID     string `json:"accountId"`
	Region        string `json:"region"`
	Partition     string `json:"partition"`
	ID            string `json:"id"`
	Arn           string `json:"arn"`
	Type          string `json:"type"`
	// The AWS resource that is associated with the activity that prompted GuardDuty to generate a finding
	// ex: //InstanceResource `json:"resource"
	Resource interface{} `json:"resource"`
	// Additional information assigned to the generated finding by GuardDuty
	// ex:  NetworkConnectionAction `json:"service"`
	Service     Service `json:"service"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Severity    float64 `json:"severity"`
	Confidence  float64 `json:"confidence,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}
