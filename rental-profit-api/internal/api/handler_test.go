package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"rental-profit-api/internal/types"
	"rental-profit-api/internal/testutil"
)

func TestMaximizeProfitHandler(t *testing.T) {

	// --- Define Test Scenarios ---
	testCases := []struct {
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

	// --- Execute Scenarios ---
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req := testutil.NewTestRequest(t, tt.requestMethod, "/maximize", tt.requestBody)
			recorder := httptest.NewRecorder() 

			MaximizeProfitHandler(recorder, req)

			if status := recorder.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
				t.Logf("Response Body: %s", recorder.Body.String()) 
			}

			if tt.expectedBodyContains != "" {
				if !strings.Contains(recorder.Body.String(), tt.expectedBodyContains) {
					t.Errorf("handler returned unexpected body: got %q want substring %q", recorder.Body.String(), tt.expectedBodyContains)
				}
			}			
		})
	}
}

func TestStatsHandler(t *testing.T) {

	// --- Define Test Scenarios ---
	testCases := []struct {
		name                 string
		requestMethod        string
		requestBody          interface{}
		expectedStatus       int
		expectedBodyContains string             
	}{
		{
			name:                 "Invalid JSON Body",
			requestMethod:        http.MethodPost,
			requestBody:          `[{"bad json":}]`,
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Invalid JSON format",
		},
		{
			name:                 "Validation Error (Missing Checkin)",
			requestMethod:        http.MethodPost,
			requestBody:          []types.BookingRequest{{RequestID: "E1", Nights: 1, SellingRate: 10, Margin: 10}}, // Missing checkin
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "check_in format error",
		},
		{
			name:                 "Validation Error (Negative Margin)",
			requestMethod:        http.MethodPost,
			requestBody:          []types.BookingRequest{{RequestID: "E1", Checkin: "2024-01-01", Nights: 1, SellingRate: 10, Margin: -10}},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "margin must be positive",
		},
		{
			name:                 "Empty Input List Success",
			requestMethod:        http.MethodPost,
			requestBody:          []types.BookingRequest{},
			expectedStatus:       http.StatusOK,
			expectedBodyContains: `{"avg_night":0,"min_night":0,"max_night":0}`,
		},
		{
			name:          "Successful Stats Calculation",
			requestMethod: http.MethodPost,
			requestBody: []types.BookingRequest{
				{RequestID: "S1", Checkin: "2024-01-01", Nights: 4, SellingRate: 100, Margin: 10},
				{RequestID: "S2", Checkin: "2024-01-06", Nights: 2, SellingRate: 150, Margin: 20}, 
				{RequestID: "S3", Checkin: "2024-01-10", Nights: 5, SellingRate: 50, Margin: 10}, 
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "Stats Calculation With One Item",
			requestMethod: http.MethodPost,
			requestBody: []types.BookingRequest{
				{RequestID: "ONE", Checkin: "2024-03-01", Nights: 3, SellingRate: 120, Margin: 25},
			},
			expectedStatus: http.StatusOK,
			expectedBodyContains: `{"avg_night":10,"min_night":10,"max_night":10}`,
		},
	}

	// --- Execute Scenarios ---
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req := testutil.NewTestRequest(t, tt.requestMethod, "/stats", tt.requestBody)
			recorder := httptest.NewRecorder() 

			StatsHandler(recorder, req)

			if status := recorder.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
				t.Logf("Response Body: %s", recorder.Body.String()) 
			}

			if tt.expectedBodyContains != "" {
				if !strings.Contains(recorder.Body.String(), tt.expectedBodyContains) {
					t.Errorf("handler returned unexpected body: got %q want substring %q", recorder.Body.String(), tt.expectedBodyContains)
				}
			}			
		})
	}

}