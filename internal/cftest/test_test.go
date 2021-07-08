package cftest

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestReadTests(t *testing.T) {
	got, err := ReadTests([]string{"testdata/error.json", "testdata/allkeys.json"})

	expected := []Test{
		{
			Filename: "testdata/error.json",
			Name:     "index",
			Event: testEvent{
				Version: "1.0",
				Context: hash{
					"eventType": "viewer-request",
				},
				Viewer: hash{
					"ip": "1.2.3.4",
				},
				Request: hash{
					"cookies": hash{},
					"headers": hash{
						"host": hash{
							"value": "www.example.com",
						},
					},
					"method":      "GET",
					"querystring": hash{},
					"uri":         "/",
				},
			},
			Error: "thrown error",
		},
		{
			Filename: "testdata/allkeys.json",
			Name:     "with all keys",
			Event: testEvent{
				Version: "1.0",
				Context: hash{
					"eventType": "viewer-response",
				},
				Viewer: hash{
					"ip": "1.1.1.1",
				},
				Request: hash{
					"cookies": hash{},
					"headers": hash{
						"host": hash{
							"value": "www.example.com",
						},
					},
					"method":      "GET",
					"querystring": hash{},
					"uri":         "/",
				},
				Response: hash{
					"statusCode":        200,
					"statusDescription": "OK",
					"cookies":           hash{},
					"headers": hash{
						"content-length": hash{
							"value": "100",
						},
						"content-type": hash{
							"value": "text/plain",
						},
						"date": hash{
							"value": "Thu, 08 Jul 2021 18:55:00 GMT",
						},
					},
				},
			},
			Output: hash{
				"request": hash{
					"cookies": hash{},
					"headers": hash{
						"host": hash{
							"value": "www.example.com",
						},
					},
					"method":      "GET",
					"querystring": hash{},
					"uri":         "/",
				},
			},
		},
	}

	if err != nil {
		t.Errorf("error: got %q, exp nil", err)
	}

	if diff := pretty.Compare(got, expected); diff != "" {
		t.Errorf("tests: (-got +want)\n%s", diff)
	}
}

func TestReadTestsNotFound(t *testing.T) {
	_, err := ReadTests([]string{"testdata/notfound.json"})
	exp := "open testdata/notfound.json: no such file or directory"
	if err == nil || err.Error() != exp {
		t.Errorf("error: got %q, exp %q", err, exp)
	}
}
