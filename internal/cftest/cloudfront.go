package cftest

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

type cloudFrontRunner struct {
	cf      *cloudfront.CloudFront
	ifMatch *string
	name    *string
	stage   *string
}

func (c cloudFrontRunner) Run(e testEvent) (*cloudfront.TestResult, error) {

	eventBytes, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	input := cloudfront.TestFunctionInput{
		EventObject: eventBytes,
		IfMatch:     c.ifMatch,
		Name:        c.name,
		Stage:       c.stage,
	}

	r, err := c.cf.TestFunction(&input)
	if err != nil {
		return nil, err
	}

	if r.TestResult.FunctionErrorMessage != nil {
		r.TestResult.FunctionErrorMessage = aws.String(strings.Replace(*r.TestResult.FunctionErrorMessage, "The CloudFront function associated with the CloudFront distribution is invalid or could not run. Error: ", "", 1))
	}

	return r.TestResult, nil
}

func NewCloudFrontRunner(name string, stage string) (Runner, error) {

	s := session.Must(session.NewSession())
	cf := cloudfront.New(s)

	r, err := cf.DescribeFunction(&cloudfront.DescribeFunctionInput{
		Name:  aws.String(name),
		Stage: aws.String(stage),
	})

	if err != nil {
		return nil, err
	}

	runner := cloudFrontRunner{
		cf:      cf,
		ifMatch: r.ETag,
		name:    &name,
		stage:   &stage,
	}
	return runner, nil
}
