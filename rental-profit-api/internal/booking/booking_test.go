package booking

import (
	"math"
	"testing"
	"time"

	"rental-profit-api/internal/types"                          
)

func parseTestDate(t *testing.T, dateStr string) time.Time {
	t.Helper()
	tm, err := time.Parse(DateLayout, dateStr)
	if err != nil {
		t.Fatalf("Failed to parse test date '%s': %v", dateStr, err)
	}
	return tm
}

func assertFloatEquals(t *testing.T, expected, actual, tolerance float64, msg string) {
	t.Helper()
	if math.Abs(expected-actual) > tolerance {
		t.Errorf("%s: Expected %f, got %f (tolerance %f)", msg, expected, actual, tolerance)
	}
}

func TestCalculateCheckout(t *testing.T) {
	tests := []struct {
		name    string
		checkin time.Time
		nights  int
		want    time.Time
	}{
		{"One night", parseTestDate(t, "2024-01-10"), 1, parseTestDate(t, "2024-01-11")},
		{"Multiple nights", parseTestDate(t, "2024-01-10"), 5, parseTestDate(t, "2024-01-15")},
		{"Across month", parseTestDate(t, "2024-01-30"), 3, parseTestDate(t, "2024-02-02")},
		{"Across year", parseTestDate(t, "2024-12-30"), 3, parseTestDate(t, "2025-01-02")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateCheckout(tt.checkin, tt.nights)
			if !got.Equal(tt.want) {
				t.Errorf("CalculateCheckout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateProfit(t *testing.T) {
	const tolerance = 1e-9
	tests := []struct {
		name        string
		sellingRate float64
		margin      float64
		nights      int
		want        float64
	}{
		{"Normal case", 100.0, 20.0, 5, 20.0}, 
		{"Fractional margin", 123.45, 15.5, 2, 19.13475},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateProfit(tt.sellingRate, tt.margin, tt.nights)
			assertFloatEquals(t, tt.want, got, tolerance, "CalculateProfit()")
		})
	}
}

func TestCalculateProfitPerNightDirect(t *testing.T) {
	const tolerance = 1e-9
	tests := []struct {
		name        string
		sellingRate float64
		margin      float64
		nights      int
		want        float64
	}{
		{"Normal case", 100.0, 20.0, 5, 4.0}, 
		{"One night", 100.0, 20.0, 1, 20.0}, 
		{"Fractional result", 150.0, 15.0, 2, 11.25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateProfitPerNight(tt.sellingRate, tt.margin, tt.nights)
			assertFloatEquals(t, tt.want, got, tolerance, "CalculateProfitPerNightDirect()")
		})
	}
}

func TestCalculateOverallStats(t *testing.T) {
	const tolerance = 1e-9
	tests := []struct {
		name     string
		bookings []Booking
		want     types.StatsResponse
	}{
		{
			name: "Single valid booking",
			bookings: []Booking{
				{SellingRate: 100, Margin: 20, Nights: 4},
			},
			want: types.StatsResponse{AvgProfitPerNight: 5.00, MinProfitPerNight: 5.00, MaxProfitPerNight: 5.00},
		},
		{
			name: "Multiple valid bookings",
			bookings: []Booking{
				{SellingRate: 100, Margin: 20, Nights: 4}, 
				{SellingRate: 200, Margin: 10, Nights: 2}, 
				{SellingRate: 50, Margin: 50, Nights: 5},  
				{SellingRate: 300, Margin: 5, Nights: 3}, 
			},
			want: types.StatsResponse{AvgProfitPerNight: 6.25, MinProfitPerNight: 5.00, MaxProfitPerNight: 10.00},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateOverallStats(tt.bookings)
			// Compare floats with tolerance
			assertFloatEquals(t, tt.want.AvgProfitPerNight, got.AvgProfitPerNight, tolerance, "AvgProfitPerNight mismatch")
			assertFloatEquals(t, tt.want.MinProfitPerNight, got.MinProfitPerNight, tolerance, "MinProfitPerNight mismatch")
			assertFloatEquals(t, tt.want.MaxProfitPerNight, got.MaxProfitPerNight, tolerance, "MaxProfitPerNight mismatch")
		})
	}
}