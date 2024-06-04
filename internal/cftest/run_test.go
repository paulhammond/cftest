package cftest

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/kylelemons/godebug/pretty"
)

type hash map[string]interface{}

func makeRequest(uri string) string {
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

func makeResponse(body hash) string {
	b, err := json.Marshal(hash{
		"response": hash{
			"statusCode":        200,
			"statusDescription": "OK",
			"headers": hash{
				"content-type": hash{
					"value": "text/plain; charset=utf-8",
				},
			},
			"body": body,
		},
	})
	if err != nil {
		panic(err)
	}
	return string(b)

}

type testRunner struct {
	result *types.TestResult
	err    error
}

func (r testRunner) Run(ctx context.Context, e testEvent) (*types.TestResult, error) {
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
			testfile: "testdata/request.json",
			runner: testRunner{result: &types.TestResult{
				ComputeUtilization: aws.String("23"),
				FunctionOutput:     aws.String(makeRequest("/")),
			}},
			exp: Result{
				Utilization: 23,
				Failure:     "",
				OK:          true,
			},
		},
		{
			name:     "ok response",
			testfile: "testdata/body_string.json",
			runner: testRunner{result: &types.TestResult{
				ComputeUtilization: aws.String("23"),
				FunctionOutput: aws.String(makeResponse(hash{
					"encoding": "text",
					"data":     "body",
				})),
			}},
			exp: Result{
				Utilization: 23,
				Failure:     "",
				OK:          true,
			},
		},
		{
			name:     "ok wildcard body",
			testfile: "testdata/body_true.json",
			runner: testRunner{result: &types.TestResult{
				ComputeUtilization: aws.String("23"),
				FunctionOutput: aws.String(makeResponse(hash{
					"encoding": "text",
					"data":     "body",
				})),
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
			runner: testRunner{result: &types.TestResult{
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
			runner: testRunner{result: &types.TestResult{
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
			testfile: "testdata/request.json",
			runner: testRunner{result: &types.TestResult{
				ComputeUtilization: aws.String("23"),
				FunctionOutput:     aws.String(makeRequest("/nomatch")),
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
			name:     "body different",
			testfile: "testdata/body_string.json",
			runner: testRunner{result: &types.TestResult{
				ComputeUtilization: aws.String("23"),
				FunctionOutput: aws.String(makeResponse(hash{
					"encoding": "text",
					"data":     "foo",
				})),
			}},
			exp: Result{
				Utilization: 23,
				Failure: `Output (-got +want):
 {
  response: {
   body: {
-   data: "foo",
+   data: "body",
    encoding: "text",
   },
   headers: {
    content-type: {
     value: "text/plain; charset=utf-8",
    },
   },
   statusCode: 200,
   statusDescription: "OK",
  },
 }`,
				OK: false,
			},
		},
		{
			name:     "body missing",
			testfile: "testdata/body_true.json",
			runner: testRunner{result: &types.TestResult{
				ComputeUtilization: aws.String("23"),
				FunctionOutput: aws.String(makeResponse(hash{
					"encoding": "text",
				})),
			}},
			exp: Result{
				Utilization: 23,
				Failure: `Output (-got +want):
 {
  response: {
   body: {
+   data: true,
    encoding: "text",
   },
   headers: {
    content-type: {
     value: "text/plain; charset=utf-8",
    },
   },
   statusCode: 200,
   statusDescription: "OK",
  },
 }`,
				OK: false,
			},
		},
		{
			name:     "test error different",
			testfile: "testdata/error.json",
			runner: testRunner{result: &types.TestResult{
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
			testfile: "testdata/request.json",
			runner:   testRunner{err: errors.New("runfail")},
			err:      errors.New("runfail"),
		},
		{
			name:     "json error",
			testfile: "testdata/request.json",
			runner: testRunner{result: &types.TestResult{
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

		result, err := RunTest(context.Background(), runner, *test)

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

func TestGetNestedBool(t *testing.T) {
	tests := []struct {
		name string
		v    any
		keys []string
		want bool
	}{
		{
			name: "true",
			v:    map[string]any{"one": map[string]any{"two": map[string]any{"three": true}}},
			keys: []string{"one", "two", "three"},
			want: true,
		},
		{
			name: "false",
			v:    map[string]any{"one": map[string]any{"two": map[string]any{"three": false}}},
			keys: []string{"one", "two", "three"},
			want: false,
		},
		{
			name: "string",
			v:    map[string]any{"one": map[string]any{"two": map[string]any{"three": "value"}}},
			keys: []string{"one", "two", "three"},
			want: false,
		},
		{
			name: "shallow",
			v:    map[string]any{"one": map[string]any{"two": map[string]any{"three": true}}},
			keys: []string{"one", "two"},
			want: false,
		},
		{
			name: "deep",
			v:    map[string]any{"one": map[string]any{"two": map[string]any{"three": true}}},
			keys: []string{"one", "two", "three", "oops"},
			want: false,
		},
		{
			name: "missing key",
			v:    map[string]any{"one": map[string]any{"two": map[string]any{"three": true}}},
			keys: []string{"one", "oops"},
			want: false,
		},
		{
			name: "missing key2",
			v:    map[string]any{"one": map[string]any{"two": map[string]any{"three": true}}},
			keys: []string{"oops"},
			want: false,
		},
		{
			name: "empty keys",
			v:    map[string]any{"one": map[string]any{"two": map[string]any{"three": true}}},
			keys: []string{},
			want: false,
		},
		{
			name: "bool object",
			v:    true,
			keys: []string{"one"},
			want: false,
		},
		{
			name: "string object",
			v:    "string",
			keys: []string{"one"},
			want: false,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			got := getNestedBool(tt.v, tt.keys)

			if got != tt.want {
				t.Errorf("got %t, want %t", got, tt.want)
			}
		})
	}
}

func TestSetNestedTrue(t *testing.T) {
	nested := map[string]any{"one": map[string]any{"two": map[string]any{"three": "value"}}}

	tests := []struct {
		name string
		v    any
		keys []string
		ok   bool
		want any
	}{
		{
			name: "ok",
			v:    map[string]any{"one": map[string]any{"two": map[string]any{"three": "value"}}},
			keys: []string{"one", "two", "three"},
			ok:   true,
			want: map[string]any{"one": map[string]any{"two": map[string]any{"three": true}}},
		},
		{
			name: "true",
			v:    map[string]any{"one": map[string]any{"two": map[string]any{"three": true}}},
			keys: []string{"one", "two", "three"},
			ok:   false,
		},
		{
			name: "false",
			v:    map[string]any{"one": map[string]any{"two": map[string]any{"three": false}}},
			keys: []string{"one", "two", "three"},
			ok:   false,
		},
		{
			name: "shallow",
			v:    nested,
			keys: []string{"one", "two"},
			ok:   false,
		},
		{
			name: "deep",
			v:    nested,
			keys: []string{"one", "two", "three", "oops"},
			ok:   false,
		},
		{
			name: "missing key",
			v:    nested,
			keys: []string{"one", "oops"},
			ok:   false,
		},
		{
			name: "missing key2",
			v:    nested,
			keys: []string{"oops"},
			ok:   false,
		},
		{
			name: "empty keys",
			v:    nested,
			keys: []string{},
			ok:   false,
		},
		{
			name: "bool object",
			v:    true,
			keys: []string{"one"},
			ok:   false,
		},
		{
			name: "string object",
			v:    "string",
			keys: []string{"one"},
			ok:   false,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			got := tt.v
			ok := setNestedTrue(got, tt.keys)

			if ok != tt.ok {
				t.Errorf("ok: got %t, want %t", ok, tt.ok)
			}
			if ok {
				if diff := pretty.Compare(got, tt.want); diff != "" {
					t.Errorf("output: (-got +want):\n%s", diff)
				}
			}
		})
	}
}
