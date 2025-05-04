package main_test // Use _test package to avoid import cycles if needed

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"testing"
	"strings"

	"rental-profit-api/internal/api"
	"rental-profit-api/internal/types"
)

type testServer struct {
	*httptest.Server
}

func startTestServer(t *testing.T) *testServer {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/maximize", api.MaximizeProfitHandler)
	mux.HandleFunc("/stats", api.StatsHandler)
	server := httptest.NewServer(mux)

	return &testServer{server}
}

func assertFloatEquals(t *testing.T, expected, actual, tolerance float64, msg string) {
	t.Helper()
	if math.Abs(expected-actual) > tolerance {
		t.Errorf("%s: Expected %f, got %f (tolerance %f)", msg, expected, actual, tolerance)
	}
}

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

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

// --- E2E Test Cases for /maximize ---

func TestMaximizeE2E(t *testing.T) {
	server := startTestServer(t)
	defer server.Close()

	// --- Define Test Scenarios ---
	testCases := []struct {
		name           			string
		requestMethod  			string	
		payload        			interface{} 
		expectedStatus 			int
		expectedResponse 		*types.MaximizeResponse
		expectedErrMsgContains 	string
	}{
		{
			name:           "Success Case - Optimal Schedule Found",
			payload: []types.BookingRequest{
				{RequestID: "B1", Checkin: "2024-01-01", Nights: 4, SellingRate: 100, Margin: 10},
				{RequestID: "B_overlap", Checkin: "2024-01-03", Nights: 3, SellingRate: 50, Margin: 10},
				{RequestID: "B3", Checkin: "2024-01-06", Nights: 2, SellingRate: 150, Margin: 20},
				{RequestID: "B4", Checkin: "2024-01-10", Nights: 3, SellingRate: 90, Margin: 15},
			},
			expectedStatus: http.StatusOK,
			expectedResponse: &types.MaximizeResponse{
				RequestIDs:  []string{"B1", "B3", "B4"},
				TotalProfit: 53.50,
				AvgNight:    7.33,
				MinNight:    2.50,
				MaxNight:    15.00,
			},
		},
		{
			name:           "Validation Error - Missing Field",
			payload:        []types.BookingRequest{{Checkin: "2024-01-01", Nights: 1, SellingRate: 10, Margin: 10}},
			expectedStatus: http.StatusBadRequest,
			expectedErrMsgContains: "request_id missing",
		},
		{
			name:           "Validation Error - Invalid Date",
			payload:        []types.BookingRequest{{RequestID: "E1", Checkin: "invalid-date", Nights: 1, SellingRate: 10, Margin: 10}},
			expectedStatus: http.StatusBadRequest,
			expectedErrMsgContains: "check_in format error",
		},
		{
			name:           "Validation Error - Non-positive Nights",
			payload:        []types.BookingRequest{{RequestID: "E2", Checkin: "2024-01-01", Nights: 0, SellingRate: 10, Margin: 10}},
			expectedStatus: http.StatusBadRequest,
			expectedErrMsgContains: "nights must be positive",
		},
		{
			name:           "Validation Error - Non-positive selling rate",
			payload: []types.BookingRequest{
				{RequestID: "Z1", Checkin: "2024-01-01", Nights: 4, SellingRate: 100, Margin: 10},
				{RequestID: "Z2", Checkin: "2024-01-06", Nights: 2, SellingRate: -50, Margin: 20},
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrMsgContains: "selling rate must be positive",
		},
		{
			name:           "Validation Error - Zero margin",
			payload: []types.BookingRequest{
				{RequestID: "Z1", Checkin: "2024-01-01", Nights: 4, SellingRate: 100, Margin: 0},
				{RequestID: "Z2", Checkin: "2024-01-06", Nights: 2, SellingRate: -50, Margin: 20},
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrMsgContains: "margin must be positive",
		},
		{
			name:           "Bad Request - Invalid JSON",
			payload:        `{"bad json": this is not valid}`,
			expectedStatus: http.StatusBadRequest,
			expectedErrMsgContains: "Invalid JSON format",
		},
		{
            name:           "Bad Request - Not An Array",
            payload:        `{"request_id":"B1","check_in":"2024-01-01","nights":4,"selling_rate":100,"margin":10}`,
            expectedStatus: http.StatusBadRequest,
            expectedErrMsgContains: "cannot unmarshal",
        },

	}

	// --- Execute Scenarios ---
	httpClient := server.Client() 

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := newHandlerTestRequest(t, testCase.requestMethod, server.URL+"/maximize", testCase.payload)

			// Send request
			resp, err := httpClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			// --- Assertions ---
			// 1. Status Code
			if resp.StatusCode != testCase.expectedStatus {
				bodyContent, _ := io.ReadAll(resp.Body)
				t.Fatalf("Expected status code %d, got %d. Body: %s", testCase.expectedStatus, resp.StatusCode, string(bodyContent))
			}

			// 2. Read Body
			bodyContent, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			// 3. Body Content
			if testCase.expectedStatus == http.StatusOK && testCase.expectedResponse != nil {
				var actualResponse types.MaximizeResponse
				err = json.Unmarshal(bodyContent, &actualResponse)
				if err != nil {
					t.Fatalf("Failed to unmarshal success response body: %v. Body: %s", err, string(bodyContent))
				}

				sort.Strings(actualResponse.RequestIDs)
				sort.Strings(testCase.expectedResponse.RequestIDs)

				if !reflect.DeepEqual(actualResponse.RequestIDs, testCase.expectedResponse.RequestIDs) {
					t.Errorf("Response RequestIDs mismatch:\n Got: %v\nWant: %v", actualResponse.RequestIDs, testCase.expectedResponse.RequestIDs)
				}
				const tolerance = 1e-2
				assertFloatEquals(t, testCase.expectedResponse.TotalProfit, actualResponse.TotalProfit, tolerance, "TotalProfit")
				assertFloatEquals(t, testCase.expectedResponse.AvgNight, actualResponse.AvgNight, tolerance, "AvgNight")
				assertFloatEquals(t, testCase.expectedResponse.MinNight, actualResponse.MinNight, tolerance, "MinNight")
				assertFloatEquals(t, testCase.expectedResponse.MaxNight, actualResponse.MaxNight, tolerance, "MaxNight")

			} else if testCase.expectedStatus != http.StatusOK && testCase.expectedErrMsgContains != "" {
				var actualError types.ErrorResponse
				err = json.Unmarshal(bodyContent, &actualError)
				if err != nil {
					if !strings.Contains(string(bodyContent), testCase.expectedErrMsgContains) {
                         t.Fatalf("Failed to unmarshal error response body: %v. Body: %s. Does not contain expected message: %q", err, string(bodyContent), testCase.expectedErrMsgContains)
                     }
				} else if !strings.Contains(actualError.Message, testCase.expectedErrMsgContains) {
					t.Errorf("Expected error message containing %q, got %q", testCase.expectedErrMsgContains, actualError.Message)
				}
			}
		})
	}
}

func TestStatsHandler(t *testing.T) {
	server := startTestServer(t)
	defer server.Close()

	// --- Define Test Scenarios ---
	testCases := []struct {
		name           string
		requestMethod  string
		payload        interface{}
		expectedStatus int
		expectedResponse *types.StatsResponse
		expectedErrMsgContains string
	}{
		{
			name:           "Method Not Allowed (GET)",
			requestMethod:  http.MethodGet,
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedErrMsgContains: "Method Not Allowed",
		},
		{
			name:           "Invalid JSON Body",
			requestMethod:  http.MethodPost,
			payload:        `[{"bad json":}]`,
			expectedStatus: http.StatusBadRequest,
			expectedErrMsgContains: "Invalid JSON format",
		},
		{
			name:           "Validation Error (Missing Nights)",
			requestMethod:  http.MethodPost,
			payload:        []types.BookingRequest{{RequestID: "V1", Checkin: "2024-01-01", SellingRate: 100, Margin: 10}}, 
			expectedStatus: http.StatusBadRequest,
			expectedErrMsgContains: "nights must be positive",
		},
		{
			name:           "Validation Error (Multiple Errors)",
			requestMethod:  http.MethodPost,
			payload: []types.BookingRequest{
				{RequestID: "E1", Checkin: "bad-date", Nights: 0, SellingRate: 100, Margin: 10},
				{RequestID: "", Checkin: "2024-01-01", Nights: 1, SellingRate: -5, Margin: -5},
			},
			expectedStatus: http.StatusBadRequest,
			expectedErrMsgContains: "check_in format error on item 0",
		},
		{
			name:           "Empty Input List Success",
			requestMethod:  http.MethodPost,
			payload:        []types.BookingRequest{},
			expectedStatus: http.StatusOK,
			expectedResponse: &types.StatsResponse{ 
				AvgProfitPerNight: 0.00, MinProfitPerNight: 0.00, MaxProfitPerNight: 0.00,
			},
		},
		{
			name:          "Successful Stats Calculation",
			requestMethod: http.MethodPost,
			payload: []types.BookingRequest{
				{RequestID: "S1", Checkin: "2024-01-01", Nights: 4, SellingRate: 100, Margin: 10}, 
				{RequestID: "S2", Checkin: "2024-01-06", Nights: 2, SellingRate: 150, Margin: 20}, 
				{RequestID: "S3", Checkin: "2024-01-10", Nights: 5, SellingRate: 50, Margin: 10}, 
			},
			expectedStatus: http.StatusOK,
			expectedResponse: &types.StatsResponse{
				AvgProfitPerNight: 6.17,
				MinProfitPerNight: 1.00,
				MaxProfitPerNight: 15.00,
			},
		},
		{
			name:          "Stats Calculation With One Item",
			requestMethod: http.MethodPost,
			payload: []types.BookingRequest{
				{RequestID: "ONE", Checkin: "2024-03-01", Nights: 3, SellingRate: 120, Margin: 25}, 
			},
			expectedStatus: http.StatusOK,
			expectedResponse: &types.StatsResponse{
				AvgProfitPerNight: 10.00, MinProfitPerNight: 10.00, MaxProfitPerNight: 10.00,
			},
		},
	}

	// --- Execute Scenarios ---
	httpClient := server.Client()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := newHandlerTestRequest(t, testCase.requestMethod, server.URL+"/stats", testCase.payload) 

			// Send request
			resp, err := httpClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			// --- Assertions ---
			if resp.StatusCode != testCase.expectedStatus {
				bodyBytes, _ := io.ReadAll(resp.Body)
				t.Fatalf("Expected status code %d, got %d. Body: %s", testCase.expectedStatus, resp.StatusCode, string(bodyBytes))
			}

			// 2. Read Body
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			bodyString := string(bodyBytes)

			// 3. Body Content
			if testCase.expectedStatus == http.StatusOK && testCase.expectedResponse != nil {
				var actualResponse types.StatsResponse
				err = json.Unmarshal(bodyBytes, &actualResponse)
				if err != nil {
					t.Fatalf("Failed to unmarshal success response body: %v. Body: %s", err, bodyString)
				}

				const tolerance = 1e-2 
				assertFloatEquals(t, testCase.expectedResponse.AvgProfitPerNight, actualResponse.AvgProfitPerNight, tolerance, "AvgProfitPerNight")
				assertFloatEquals(t, testCase.expectedResponse.MinProfitPerNight, actualResponse.MinProfitPerNight, tolerance, "MinProfitPerNight")
				assertFloatEquals(t, testCase.expectedResponse.MaxProfitPerNight, actualResponse.MaxProfitPerNight, tolerance, "MaxProfitPerNight")

			} else if testCase.expectedStatus != http.StatusOK && testCase.expectedErrMsgContains != "" {
				var actualError types.ErrorResponse
				err = json.Unmarshal(bodyBytes, &actualError)
				if err != nil { 
					if !strings.Contains(bodyString, testCase.expectedErrMsgContains) {
						t.Errorf("Expected error response containing %q, but unmarshal failed and raw body %q did not contain it", testCase.expectedErrMsgContains, bodyString)
					}
				} else if !strings.Contains(actualError.Message, testCase.expectedErrMsgContains) { 
					t.Errorf("Expected error message containing %q, got %q", testCase.expectedErrMsgContains, actualError.Message)
				}
			}
		})
	}
}