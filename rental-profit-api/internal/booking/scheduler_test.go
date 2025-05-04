package booking

import (
	"reflect"
	"sort"
	"testing"
)

func newTestBooking(t *testing.T, id string, checkinStr string, nights int, rate, margin float64) Booking {
	t.Helper()
	b := Booking{
		RequestID:   id,
		Checkin:     parseTestDate(t, checkinStr),
		Nights:      nights,
		SellingRate: rate,
		Margin:      margin,
	}
	if nights > 0 {
		b.Checkout = CalculateCheckout(b.Checkin, b.Nights)
		b.Profit = CalculateProfit(b.SellingRate, b.Margin, b.Nights)
	}
	return b
}

func sortBookingsByCheckout(bookings []Booking) {
	sort.Slice(bookings, func(i, j int) bool {
		if bookings[i].Checkout.Equal(bookings[j].Checkout) {
			return bookings[i].Checkin.Before(bookings[j].Checkin)
		}
		return bookings[i].Checkout.Before(bookings[j].Checkout)
	})
}

func assertScheduleResult(t *testing.T, expected ScheduleResult, actual ScheduleResult) {
	t.Helper()
	const tolerance = 1e-9

	// 1. Check Schedule Content (IDs and Order)
	if len(expected.OptimalSchedule) != len(actual.OptimalSchedule) {
		t.Errorf("Expected schedule length %d, got %d. Expected: %v, Got: %v",
			len(expected.OptimalSchedule), len(actual.OptimalSchedule), expected.OptimalSchedule, actual.OptimalSchedule)
		// Stop further comparison if lengths differ significantly
		return
	}
	expectedIDs := make([]string, len(expected.OptimalSchedule))
	actualIDs := make([]string, len(actual.OptimalSchedule))
	for i := range expected.OptimalSchedule {
		expectedIDs[i] = expected.OptimalSchedule[i].RequestID
	}
	for i := range actual.OptimalSchedule {
		actualIDs[i] = actual.OptimalSchedule[i].RequestID
	}
	if !reflect.DeepEqual(expectedIDs, actualIDs) {
		t.Errorf("Booking schedule mismatch:\nExpected IDs: %v\nGot IDs:      %v", expectedIDs, actualIDs)
	}

	// 2. Check Total Profit
	assertFloatEquals(t, expected.TotalProfit, actual.TotalProfit, tolerance, "TotalProfit mismatch")

	// 3. Check PPN Stats
	assertFloatEquals(t, expected.AvgProfitPerNight, actual.AvgProfitPerNight, tolerance, "AvgProfitPerNight mismatch")
	assertFloatEquals(t, expected.MinProfitPerNight, actual.MinProfitPerNight, tolerance, "MinProfitPerNight mismatch")
	assertFloatEquals(t, expected.MaxProfitPerNight, actual.MaxProfitPerNight, tolerance, "MaxProfitPerNight mismatch")
}

func TestFindLatestCompatibleBinarySearch(t *testing.T) {
	bookings := []Booking{
		newTestBooking(t, "B1", "2024-01-01", 4, 100, 10),
		newTestBooking(t, "B2", "2024-01-06", 3, 100, 10),
		newTestBooking(t, "B3", "2024-01-05", 5, 100, 10),
		newTestBooking(t, "B4", "2024-01-10", 2, 100, 10),
		newTestBooking(t, "B5", "2024-01-11", 3, 100, 10),
	}
	sortBookingsByCheckout(bookings)

	tests := []struct {
		name          string
		targetIndex   int
		targetBooking Booking
		want          int
	}{
		{
			name:          "No compatible (target is first)",
			targetIndex:   0,
			targetBooking: bookings[0],
			want:          -1,
		},
		{
			name:          "One compatible (B2 vs B1)",
			targetIndex:   1,
			targetBooking: bookings[1],
			want:          0,
		},
		{
			name:          "Adjacent (B3 vs B2)",
			targetIndex:   2,
			targetBooking: bookings[2],
			want:          0,
		},
		{
			name:          "Multiple compatible, find latest (B4 vs B1, B2, B3)",
			targetIndex:   3,
			targetBooking: bookings[3],
			want:          2,
		},
		{
			name:          "Overlap with immediate predecessor (B5 vs B4)",
			targetIndex:   4,
			targetBooking: bookings[4],
			want:          2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findLatestCompatibleBinarySearch(bookings, tt.targetIndex)
			if got != tt.want {
				t.Errorf("findLatestCompatibleBinarySearch() for booking %s (index %d) = %v, want %v", tt.targetBooking.RequestID, tt.targetIndex, got, tt.want)
			}
		})
	}
}

func TestFindMaxProfit(t *testing.T) {
	bookingSet1 := []Booking{
		newTestBooking(t, "B1", "2024-01-01", 4, 100, 10),
		newTestBooking(t, "B2", "2024-01-06", 2, 150, 20), 
	}
	expectedResult1 := ScheduleResult{
		OptimalSchedule:   []Booking{bookingSet1[0], bookingSet1[1]},
		TotalProfit:       40.0, 
		AvgProfitPerNight: 8.75, 
		MinProfitPerNight: 2.5,
		MaxProfitPerNight: 15.0,
	}

	bookingSet2 := []Booking{ 
		newTestBooking(t, "B1", "2024-01-01", 5, 100, 10), 
		newTestBooking(t, "B2", "2024-01-04", 4, 150, 20), 
	}
	expectedResult2 := ScheduleResult{
		OptimalSchedule:   []Booking{bookingSet2[1]}, 
		TotalProfit:       30.0,
		AvgProfitPerNight: 7.5,
		MinProfitPerNight: 7.5,
		MaxProfitPerNight: 7.5,
	}

	bookingSet3 := []Booking{ 
		newTestBooking(t, "B1", "2024-01-01", 4, 100, 20), 
		newTestBooking(t, "B2", "2024-01-03", 5, 100, 30), 
		newTestBooking(t, "B3", "2024-01-06", 2, 100, 25), 
	}
	expectedResult3 := ScheduleResult{
		OptimalSchedule:   []Booking{bookingSet3[0], bookingSet3[2]},
		TotalProfit:       45.0,  
		AvgProfitPerNight: 8.75, 
		MinProfitPerNight: 5.0,
		MaxProfitPerNight: 12.5,
	}
    
	bookingSet5 := []Booking{ 
		newTestBooking(t, "S1", "2024-02-01", 3, 90, 10),
	}
    expectedResult5 := ScheduleResult{
		OptimalSchedule:   []Booking{bookingSet5[0]},
		TotalProfit:       9.0,
		AvgProfitPerNight: 3.0,
		MinProfitPerNight: 3.0,
		MaxProfitPerNight: 3.0,
	}

	tests := []struct {
		name           string
		bookings       []Booking
		expectedResult ScheduleResult
	}{
		{"Empty input", []Booking{}, ScheduleResult{OptimalSchedule: []Booking{}, TotalProfit: 0, AvgProfitPerNight: 0, MinProfitPerNight: 0, MaxProfitPerNight: 0}},
        {"Single booking", bookingSet5, expectedResult5},
		{"Two non-overlapping", bookingSet1, expectedResult1},
		{"Two overlapping pick higher profit", bookingSet2, expectedResult2},
		{"Overlapping pick combo", bookingSet3, expectedResult3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := FindMaxProfit(tt.bookings)

			sort.Slice(tt.expectedResult.OptimalSchedule, func(i, j int) bool {
				return tt.expectedResult.OptimalSchedule[i].Checkin.Before(tt.expectedResult.OptimalSchedule[j].Checkin)
			})

			assertScheduleResult(t, tt.expectedResult, gotResult)
		})
	}
}