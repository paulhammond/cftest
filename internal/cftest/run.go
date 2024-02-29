package cftest

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/kylelemons/godebug/pretty"
)

type Result struct {
	OK          bool
	Utilization int
	Failure     string
}

type Runner interface {
	Run(ctx context.Context, e testEvent) (*types.TestResult, error)
	Name() string
}

func RunTest(ctx context.Context, runner Runner, test Test) (Result, error) {

	result := Result{}
	failures := []string{}

	testResult, err := runner.Run(ctx, test.Event)
	if err != nil {
		return result, err
	}

	if testResult.ComputeUtilization != nil {
		result.Utilization, err = strconv.Atoi(aws.ToString(testResult.ComputeUtilization))
		if err != nil {
			return result, err
		}
	}

	if gotError := aws.ToString(testResult.FunctionErrorMessage); gotError != test.Error {
		failures = append(failures, fmt.Sprintf("Error (-got +want):\n%s", pretty.Compare(gotError, test.Error)))
	}

	var output interface{}
	if testResult.FunctionOutput != nil {
		err := json.Unmarshal([]byte(*testResult.FunctionOutput), &output)
		if err != nil {
			return result, fmt.Errorf("JSON decode error: %w", err)
		}
	}
	if t, ok := output.(map[string]interface{}); ok && len(t) == 0 {
		output = nil
	}

	if diff := pretty.Compare(output, test.Output); diff != "" {
		failures = append(failures, fmt.Sprintf("Output (-got +want):\n%s", diff))
	}

	if len(failures) > 0 {
		result.Failure = strings.Join(failures, "\n")
	} else {
		result.OK = true
	}
	return result, nil

}
