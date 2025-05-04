package api

import (
	"fmt"
	"math"
	"errors"
	"net/http"
	"encoding/json"
	"time"

	"rental-profit-api/internal/booking"
	"rental-profit-api/internal/types"
)

func MaximizeProfitHandler(w http.ResponseWriter, r *http.Request) {
	var bookingRequest []types.BookingRequest
	err := json.NewDecoder(r.Body).Decode(&bookingRequest)
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON format: %v", err))
		return
	}
	defer r.Body.Close()

	// Validate the request content and format
	domainBookings, err := validateAndMapBookings(bookingRequest) 
	if err != nil {
		if errors.Is(err, ErrValidation) {
			respondError(w, http.StatusBadRequest, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// An empty request is valid, but the response will also be empty
	if len(domainBookings) == 0 {
		respondJSON(w, http.StatusOK, types.MaximizeResponse{
			RequestIDs:  []string{},
			TotalProfit: 0.0,
			AvgNight:    0.0,
			MinNight:    0.0,
			MaxNight:    0.0,
		})
		return
	}

	// Execute business logic
	var scheduleResult booking.ScheduleResult
	var panicErr any
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = r
			}
		}()
		scheduleResult = booking.FindMaxProfit(domainBookings)
	}()

	if panicErr != nil {
		respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	requestIDs := make([]string, len(scheduleResult.OptimalSchedule))
	for i, b := range scheduleResult.OptimalSchedule {
		requestIDs[i] = b.RequestID
	}

	// Apply Rounding for Presentation
	totalProfitRounded := math.Round(scheduleResult.TotalProfit*100) / 100
	avgNightRounded := math.Round(scheduleResult.AvgProfitPerNight*100) / 100
	minNightRounded := math.Round(scheduleResult.MinProfitPerNight*100) / 100
	maxNightRounded := math.Round(scheduleResult.MaxProfitPerNight*100) / 100

	response := types.MaximizeResponse{
		RequestIDs:  requestIDs,
		TotalProfit: totalProfitRounded,
		AvgNight:    avgNightRounded,
		MinNight:    minNightRounded,
		MaxNight:    maxNightRounded,
	}

	respondJSON(w, http.StatusOK, response)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var bookingRequest []types.BookingRequest 
	err := json.NewDecoder(r.Body).Decode(&bookingRequest)
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON format: %v", err))
		return
	}
	defer r.Body.Close()

	// Validate the request content and format
	domainBookings, err := validateAndMapBookings(bookingRequest)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			respondError(w, http.StatusBadRequest, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	// An empty request is valid, but the response will also be empty
	if len(domainBookings) == 0 {
		respondJSON(w, http.StatusOK, types.StatsResponse{
			AvgProfitPerNight: 0.0,
			MinProfitPerNight: 0.0,
			MaxProfitPerNight: 0.0,
		})
		return
	}

	// Execute business logic
	var statsResult types.StatsResponse
	var panicErr any // Use any for panic value
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = r
			}
		}()
		statsResult = booking.CalculateOverallStats(domainBookings)
	}()

	if panicErr != nil {
		respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	respondJSON(w, http.StatusOK, statsResult)
}

var ErrValidation = errors.New("validation error")

func validateAndMapBookings(requestItems []types.BookingRequest) ([]booking.Booking, error) {
	domainBookings := make([]booking.Booking, 0, len(requestItems))
	for i, item := range requestItems {
		if item.RequestID == "" {
			return nil, fmt.Errorf("%w: request_id missing on item %d", ErrValidation, i)
		}
		checkinDate, err := time.Parse(booking.DateLayout, item.Checkin)
		if err != nil {
			return nil, fmt.Errorf("%w: check_in format error on item %d: %w", ErrValidation, i, err)
		}
		if item.Nights <= 0 {
			return nil, fmt.Errorf("%w: nights must be positive on item %d", ErrValidation, i)
		}
		if item.SellingRate <= 0 {
			return nil, fmt.Errorf("%w: selling rate must be positive on item %d", ErrValidation, i)
		}
		if item.Margin <= 0 {
			return nil, fmt.Errorf("%w: margin must be positive on item %d", ErrValidation, i)
		}
		domainBookings = append(domainBookings, booking.Booking{
			RequestID:   item.RequestID,
			Checkin:     checkinDate,
			Nights:      item.Nights,
			SellingRate: item.SellingRate,
			Margin:      item.Margin,
		})
	}
	return domainBookings, nil
}