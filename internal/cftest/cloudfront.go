package cftest

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

type cloudFrontRunner struct {
	cf      *cloudfront.Client
	ifMatch *string
	name    string
	stage   types.FunctionStage
}

func (c cloudFrontRunner) Name() string {
	return c.name + " " + string(c.stage)
}

func (c cloudFrontRunner) Run(ctx context.Context, e testEvent) (*types.TestResult, error) {

	eventBytes, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	input := cloudfront.TestFunctionInput{
		EventObject: eventBytes,
		IfMatch:     c.ifMatch,
		Name:        &c.name,
		Stage:       c.stage,
	}

	r, err := c.cf.TestFunction(ctx, &input)
	if err != nil {
		return nil, err
	}

	if r.TestResult.FunctionErrorMessage != nil {
		r.TestResult.FunctionErrorMessage = aws.String(strings.Replace(*r.TestResult.FunctionErrorMessage, "The CloudFront function associated with the CloudFront distribution is invalid or could not run. Error: ", "", 1))
	}

	return r.TestResult, nil
}

func NewCloudFrontRunner(ctx context.Context, name string, stage string) (Runner, error) {

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	functionStage := types.FunctionStage(stage)

	cf := cloudfront.NewFromConfig(cfg)

	r, err := cf.DescribeFunction(ctx, &cloudfront.DescribeFunctionInput{
		Name:  aws.String(name),
		Stage: functionStage,
	})

	if err != nil {
		return nil, err
	}

	runner := cloudFrontRunner{
		cf:      cf,
		ifMatch: r.ETag,
		name:    name,
		stage:   functionStage,
	}
	return runner, nil
}
