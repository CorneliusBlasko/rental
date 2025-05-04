package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"rental-profit-api/internal/types"
)


func newHandlerTestRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()
	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(b)
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

func TestMaximizeProfitHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestMethod  string
		requestBody    interface{}
		expectedStatus int
		expectedBodyContains string
	}{
		{
			name:           "Invalid JSON Body",
			requestMethod:  http.MethodPost,
			requestBody:    `{"bad json":}`, 
			expectedStatus: http.StatusBadRequest,
			expectedBodyContains: "Invalid JSON format",
		},
		{
			name:           "Validation Error (Missing RequestID)",
			requestMethod:  http.MethodPost,
			requestBody:    []types.BookingRequest{{Checkin: "2024-01-01", Nights: 1, SellingRate: 10, Margin: 10}},
			expectedStatus: http.StatusBadRequest,
			expectedBodyContains: "request_id missing",
		},
        {
			name:           "Validation Error (Zero Nights)",
			requestMethod:  http.MethodPost,
			requestBody:    []types.BookingRequest{{RequestID: "A", Checkin: "2024-01-01", Nights: 0, SellingRate: 10, Margin: 10}},
			expectedStatus: http.StatusBadRequest,
			expectedBodyContains: "nights must be positive",
		},
		{
			name:           "Empty Input List Success",
			requestMethod:  http.MethodPost,
			requestBody:    []types.BookingRequest{},
			expectedStatus: http.StatusOK,
			expectedBodyContains: `"request_ids":[]`,
		},
		{
			name:          "Successful Calculation (Basic)",
			requestMethod: http.MethodPost,
			requestBody: []types.BookingRequest{
				{RequestID: "B1", Checkin: "2024-01-01", Nights: 4, SellingRate: 100, Margin: 10},
				{RequestID: "B2", Checkin: "2024-01-06", Nights: 2, SellingRate: 150, Margin: 20},
			},
			expectedStatus: http.StatusOK,
			expectedBodyContains: `"request_ids":["B1","B2"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newHandlerTestRequest(t, tt.requestMethod, "/maximize", tt.requestBody)
			rr := httptest.NewRecorder() 

			MaximizeProfitHandler(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
				t.Logf("Response Body: %s", rr.Body.String()) 
			}

			if tt.expectedBodyContains != "" {
				if !strings.Contains(rr.Body.String(), tt.expectedBodyContains) {
					t.Errorf("handler returned unexpected body: got %q want substring %q", rr.Body.String(), tt.expectedBodyContains)
				}
			}			
		})
	}
}

// Add similar tests for StatsHandler
func TestStatsHandler(t *testing.T) {

}