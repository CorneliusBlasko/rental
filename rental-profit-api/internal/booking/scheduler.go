package booking

import (
	"math"
	"slices"
)

func findLatestCompatibleBinarySearch(bookings []Booking, i int) int {
	targetCheckin := bookings[i].Checkin
	low, high := 0, i-1
	latestCompatible := -1

	for low <= high {
		mid := low + (high-low)/2
		if !bookings[mid].Checkout.After(targetCheckin) {
			latestCompatible = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return latestCompatible
}

func FindMaxProfit(inputBookings []Booking) ScheduleResult {
	bookingsLength := len(inputBookings)

	result := ScheduleResult{ // Initialize result struct
		OptimalSchedule: []Booking{},
		TotalProfit:       0.0,
		AvgProfitPerNight: 0.0,
		MinProfitPerNight: 0.0,
		MaxProfitPerNight: 0.0,
	}

	if bookingsLength == 0 {
		return result
	}

	// 1.- Calculate the checkout date and profit for each booking
	bookings := make([]Booking, bookingsLength)
	for i, booking := range inputBookings {
		bookings[i] = booking
		bookings[i].Checkout = CalculateCheckout(booking.Checkin, booking.Nights)
		bookings[i].Profit = CalculateProfit(booking.SellingRate, booking.Margin, booking.Nights)
	}

	// 2.- Sort bookings by Checkout time
	slices.SortFunc(bookings, func(a, b Booking) int {
		checkoutComparision := a.Checkout.Compare(b.Checkout)
		if checkoutComparision != 0 {
			return checkoutComparision
		}
		return a.Checkin.Compare(b.Checkin)
	})

	// 3.- Calculate the latest compatible predecessor for each booking using binary search
	latestCompatiblePredecessors := make([]int, bookingsLength)
	for i := range bookingsLength {
		latestCompatiblePredecessors[i] = findLatestCompatibleBinarySearch(bookings, i)
	}

	// 4.- Calculate max profit up to index i
	dp := make([]float64, bookingsLength)
	if bookingsLength > 0 {
		dp[0] = math.Max(0, bookings[0].Profit)
	}

	for i := 1; i < bookingsLength; i++ {
		profit_of_i := bookings[i].Profit
		compatibleProfit := 0.0
		if latestCompatiblePredecessors[i] != -1 {
			compatibleProfit = dp[latestCompatiblePredecessors[i]]
		}
		profitIncluding_i := profit_of_i + compatibleProfit
		profitExcluding_i := dp[i-1]

		dp[i] = math.Max(profitIncluding_i, profitExcluding_i)
	}

	// 5.- Find the overall maximum profit
	maxProfit := 0.0
	if bookingsLength > 0 {
		maxProfit = dp[bookingsLength-1]
	}
	if maxProfit <= 0 {
		return result
	}

	// 6.- Reconstruct the optimal schedule by backtracking through DP decisions
	optimalSchedule := []Booking{}
	i := bookingsLength - 1
	currentExpectedProfit := maxProfit // Start from the end result

	// Tolerance for float comparison
	const tolerance = 1e-9

	for i >= 0 {
		if i == 0 {
			if bookings[0].Profit > 0 && math.Abs(dp[0]-bookings[0].Profit) < tolerance {
				optimalSchedule = append(optimalSchedule, bookings[0])
			}
			break
		}

		profitExcluding_i := dp[i-1]

		if math.Abs(currentExpectedProfit-profitExcluding_i) < tolerance {
			i--
			currentExpectedProfit = dp[i]
		} else {
			if bookings[i].Profit > 0 {
				optimalSchedule = append(optimalSchedule, bookings[i])
			}

			prevCompatibleIndex := latestCompatiblePredecessors[i]
			if prevCompatibleIndex != -1 {
				currentExpectedProfit = dp[prevCompatibleIndex]
			} else {
				currentExpectedProfit = 0
			}
			i = prevCompatibleIndex
			if i == -1 {
				break
			}
		}
	}

	// 7.- Reversing the schedule because it was created backwards from latest checkout
	slices.Reverse(optimalSchedule)
	result.OptimalSchedule = optimalSchedule

	// 8.- Calculate the profits
	var totalProfit float64
	var totalProfitPerNight float64
	scheduleLen := len(result.OptimalSchedule)
	isFirst := true

	for _, boooking := range result.OptimalSchedule {
		totalProfit += boooking.Profit 

		// This should never happen, but we act defensively to avoid division by zero
		profitPerNight := 0.0
		if boooking.Nights > 0 {
			profitPerNight = boooking.Profit / float64(boooking.Nights)
		}

		totalProfitPerNight += profitPerNight

		if isFirst {
			result.MinProfitPerNight = profitPerNight
			result.MaxProfitPerNight = profitPerNight
			isFirst = false
		} else {
			if profitPerNight < result.MinProfitPerNight {
				result.MinProfitPerNight = profitPerNight
			}
			if profitPerNight > result.MaxProfitPerNight {
				result.MaxProfitPerNight = profitPerNight
			}
		}
	}

	result.TotalProfit = totalProfit
	
	if scheduleLen > 0 {
		result.AvgProfitPerNight = totalProfitPerNight / float64(scheduleLen)
	}

	return result
}
