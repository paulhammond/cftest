package cftest

import (
	"encoding/json"
	"os"
)

type Test struct {
	Filename string      `json:"-"`
	Event    testEvent   `json:"event"`
	Output   interface{} `json:"output"`
	Error    string      `json:"error"`
}

type testEvent struct {
	Version  string      `json:"version"`
	Context  interface{} `json:"context"`
	Viewer   interface{} `json:"viewer"`
	Request  interface{} `json:"request,omitempty"`
	Response interface{} `json:"response,omitempty"`
}

func ReadTests(files []string) ([]Test, error) {

	tests := []Test{}
	for _, f := range files {
		t, err := readTest(f)
		if err != nil {
			return nil, err
		}
		tests = append(tests, *t)
	}
	return tests, nil
}

func readTest(path string) (*Test, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var t = Test{
		Filename: path,
		Event: testEvent{
			Version: "1.0",
			Context: map[string]interface{}{
				"eventType": "viewer-request",
			},
			Viewer: map[string]interface{}{
				"ip": "1.2.3.4",
			},
		},
	}
	err = json.Unmarshal(f, &t)
	return &t, err
}
