package booking

import (
	"time"
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

func CalculateProfitPerNight(totalProfit float64, nights int) float64 {
	if nights <= 0 {
		return 0
	}
	return totalProfit / float64(nights)
}