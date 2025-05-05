package testutil 

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func NewTestRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var reqBody *bytes.Buffer
	if body != nil {
		marshalledBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(marshalledBody)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}