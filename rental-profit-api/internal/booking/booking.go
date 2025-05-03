package booking

import (
	"time"
	"math"

	"rental-profit-api/internal/types"
)

const DateLayout = "2006-01-02"

type Booking struct {
	RequestID   string    
	Checkin     time.Time 
	Nights      int
	SellingRate float64
	Margin      float64
	Checkout time.Time
	Profit   float64
}

func CalculateCheckout(checkin time.Time, nights int) time.Time {
	if nights <= 0 {
		return checkin
	}
	return checkin.AddDate(0, 0, nights)
}

func CalculateProfit(sellingRate, margin float64, nights int) float64 {
	if nights <= 0 || margin < 0 || sellingRate < 0 {
		return 0
	}
	return sellingRate * (margin / 100.0)
}

func CalculateProfitPerNight(sellingRate, margin float64, nights int) float64 {
	if nights <= 0 {
		return 0
	}
	totalProfit := sellingRate * (margin / 100.0)
	return totalProfit / float64(nights)
}

type ScheduleResult struct {
	OptimalSchedule   []Booking
	TotalProfit       float64
	AvgProfitPerNight float64 
	MinProfitPerNight float64 
	MaxProfitPerNight float64
}

func CalculateOverallStats(bookings []Booking) types.StatsResponse {
	response := types.StatsResponse{
		MinProfitPerNight: 0.0,
		MaxProfitPerNight: 0.0,
		AvgProfitPerNight: 0.0,
	}

	var sumProfitPerNight float64
	var validCount int = 0 
	firstValid := true

	for _, b := range bookings {
		profitPerNight := CalculateProfitPerNight(b.SellingRate, b.Margin, b.Nights)

		sumProfitPerNight += profitPerNight
		validCount++ 

		if firstValid {
			response.MinProfitPerNight = profitPerNight
			response.MaxProfitPerNight = profitPerNight
			firstValid = false
		} else {
			if profitPerNight < response.MinProfitPerNight {
				response.MinProfitPerNight = profitPerNight
			}
			if profitPerNight > response.MaxProfitPerNight {
				response.MaxProfitPerNight = profitPerNight
			}
		}
	}

	if validCount > 0 {
		response.AvgProfitPerNight = sumProfitPerNight / float64(validCount)
	} else {
		response.MinProfitPerNight = 0.0
		response.MaxProfitPerNight = 0.0
        response.AvgProfitPerNight = 0.0
	}
	
	response.AvgProfitPerNight = math.Round(response.AvgProfitPerNight*100) / 100
	response.MinProfitPerNight = math.Round(response.MinProfitPerNight*100) / 100
	response.MaxProfitPerNight = math.Round(response.MaxProfitPerNight*100) / 100

	return response

}