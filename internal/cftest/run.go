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

	bodyDataKey := []string{"response", "body", "data"}
	if getNestedBool(test.Output, bodyDataKey) {
		setNestedTrue(output, bodyDataKey)
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

func getNestedBool(v any, keys []string) bool {
	if len(keys) == 0 {
		return false
	}
	for _, k := range keys {
		m, ok := v.(map[string]any)
		if !ok {
			return false
		}

		v, ok = m[k]
		if !ok {
			return false
		}
	}

	s, ok := v.(bool)
	if !ok {
		return false
	}
	return s
}

func setNestedTrue(v any, keys []string) bool {
	if len(keys) == 0 {
		return false
	}
	for i, k := range keys {
		m, ok := v.(map[string]any)
		if !ok {
			return false
		}

		v, ok = m[k]
		if !ok {
			return false
		}

		if i == len(keys)-1 {
			_, ok := v.(string)
			if !ok {
				return false
			}
			m[k] = true
			return true
		}
	}
	panic("unreachable")
}
