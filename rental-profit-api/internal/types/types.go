package types

type BookingRequest struct {
	RequestID   string  `json:"request_id"`
	Checkin     string  `json:"check_in"` 
	Nights      int     `json:"nights"`
	SellingRate float64 `json:"selling_rate"`
	Margin      float64 `json:"margin"`
}

type MaximizeResponse struct {
	RequestIDs []string `json:"request_ids"`
	TotalProfit float64 `json:"total_profit"`
	AvgNight    float64 `json:"avg_night"` 
	MinNight    float64 `json:"min_night"`
	MaxNight    float64 `json:"max_night"`
}

type StatsResponse struct {
	AvgProfitPerNight float64 `json:"avg_night"`
	MinProfitPerNight float64 `json:"min_night"`
	MaxProfitPerNight float64 `json:"max_night"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type ProfitStats struct {
	TotalProfit float64
	AvgNight    float64
	MinNight    float64
	MaxNight    float64
	RequestIDs  []string
}