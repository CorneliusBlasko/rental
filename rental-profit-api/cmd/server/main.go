package main

import (
	"log/slog"
	"net/http"
	"os"

	"rental-profit-api/internal/api"
)

func main() {


	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	slog.SetDefault(logger) 

	slog.Info("Initializing server...")

	// --- HTTP Route Registration ---
	http.HandleFunc("/maximize", api.MaximizeProfitHandler)
	slog.Info("Registered handler for endpoint", "path", "/maximize")

	// --- Server Configuration ---
	port := "8080"
	addr := ":" + port
	slog.Info("Server starting", "address", addr)

	// --- Start Server ---
	err := http.ListenAndServe(addr, nil)

	// --- Error Handling ---
	if err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}