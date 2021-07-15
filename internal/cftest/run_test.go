package cftest

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

type hash map[string]interface{}

func makeResult(uri string) string {
	b, err := json.Marshal(hash{
		"request": hash{
			"method":      "GET",
			"uri":         uri,
			"querystring": hash{},
			"headers": hash{
				"host": hash{
					"value": "www.example.com",
				},
			},
			"cookies": hash{},
		},
	})
	if err != nil {
		panic(err)
	}
	return string(b)
}

type testRunner struct {
	result *cloudfront.TestResult
	err    error
}

func (r testRunner) Run(e testEvent) (*cloudfront.TestResult, error) {
	return r.result, r.err
}

func (r testRunner) Name() string {
	return "testRunner"
}

func TestRunTest(t *testing.T) {

	tests := []struct {
		name     string
		testfile string
		runner   testRunner
		exp      Result
		err      error
	}{
		{
			name:     "ok",
			testfile: "testdata/index.json",
			runner: testRunner{result: &cloudfront.TestResult{
				ComputeUtilization: aws.String("23"),
				FunctionOutput:     aws.String(makeResult("/")),
			}},
			exp: Result{
				Utilization: 23,
				Failure:     "",
				OK:          true,
			},
		},
		{
			name:     "expected test error",
			testfile: "testdata/error.json",
			runner: testRunner{result: &cloudfront.TestResult{
				ComputeUtilization:   aws.String("23"),
				FunctionErrorMessage: aws.String("thrown error"),
			}},
			exp: Result{
				Utilization: 23,
				Failure:     "",
				OK:          true,
			},
		},
		{
			name:     "expected test error with empty output",
			testfile: "testdata/error.json",
			runner: testRunner{result: &cloudfront.TestResult{
				ComputeUtilization:   aws.String("23"),
				FunctionErrorMessage: aws.String("thrown error"),
				FunctionOutput:       aws.String("{}"),
			}},
			exp: Result{
				Utilization: 23,
				Failure:     "",
				OK:          true,
			},
		},
		{
			name:     "output different",
			testfile: "testdata/index.json",
			runner: testRunner{result: &cloudfront.TestResult{
				ComputeUtilization: aws.String("23"),
				FunctionOutput:     aws.String(makeResult("/nomatch")),
			}},
			exp: Result{
				Utilization: 23,
				Failure: `Output (-got +want):
 {
  request: {
   cookies: {
   },
   headers: {
    host: {
     value: "www.example.com",
    },
   },
   method: "GET",
   querystring: {
   },
-  uri: "/nomatch",
+  uri: "/",
  },
 }`,
				OK: false,
			},
		},
		{
			name:     "test error different",
			testfile: "testdata/error.json",
			runner: testRunner{result: &cloudfront.TestResult{
				ComputeUtilization:   aws.String("23"),
				FunctionErrorMessage: aws.String("other error"),
			}},
			exp: Result{
				Utilization: 23,
				Failure: `Error (-got +want):
-"other error"
+"thrown error"`,
				OK: false,
			},
		},
		{
			name:     "runner error",
			testfile: "testdata/index.json",
			runner:   testRunner{err: errors.New("runfail")},
			err:      errors.New("runfail"),
		},
		{
			name:     "json error",
			testfile: "testdata/index.json",
			runner: testRunner{result: &cloudfront.TestResult{
				FunctionOutput: aws.String("/"),
			}},
			err: errors.New("JSON decode error: invalid character '/' looking for beginning of value"),
		},
	}

	for _, tt := range tests {
		runner := tt.runner
		test, err := readTest(tt.testfile)
		if err != nil {
			panic(err)
		}

		result, err := RunTest(runner, *test)

		checkErrorEqual(t, tt.name, err, tt.err)

		if tt.err == nil {
			if !reflect.DeepEqual(tt.exp, result) {
				t.Errorf("%s: result:\nexp: %#v\ngot: %#v\n", tt.name, tt.exp, result)
			}
		}

	}

}

func checkErrorEqual(tb testing.TB, name string, a, b error) {
	tb.Helper()
	if a == nil && b == nil || a != nil && b != nil && a.Error() == b.Error() {
		return
	}
	tb.Fatalf("%s: error: exp %q got %q", name, a, b)
}
