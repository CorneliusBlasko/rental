# Rental Profit Maximizer API

## Overview

This application provides an API to help manage apartment bookings by maximizing potential profit. It accepts a list of booking requests, each with check-in date, duration (nights), selling rate, and margin. The core `/maximize` endpoint uses a Dynamic Programming algorithm (specifically, Weighted Interval Scheduling) to determine the single best combination of non-overlapping bookings that yields the highest total profit. It also provides a `/stats` endpoint to calculate aggregate profit-per-night statistics across all submitted (valid) booking requests.

## Getting Started (Docker)

1. **Prerequisites:** Ensure you have Git, Docker, and Docker Compose installed.
2. **Clone:** Clone the repository to your local machine:
    ```bash
    git clone https://github.com/CorneliusBlasko/rental.git
    cd rental
    ```
3. **Run:** Build and run the application using Docker Compose:
    ```bash
    docker-compose up --build -d
    ```
    The API will be available at `http://localhost:8080`.
    
4. **Test:** You can test the `/maximize`  and the `/stats` endpoints using `curl`:
    ```bash
    # Maximize endpoint curl command
    curl -X POST -H 'Content-Type: application/json' -d '[{"request_id":"B1","check_in":"2024-01-01","nights":4,"selling_rate":100,"margin":10},{"request_id":"B3","check_in":"2024-01-06","nights":2,"selling_rate":150,"margin":20}]' http://localhost:8080/maximize

    # Stats endpoint curl command:
    curl -X POST -H 'Content-Type: application/json' -d '[{"request_id":"B1","check_in":"2024-01-01","nights":4,"selling_rate":100,"margin":10},{"request_id":"B3","check_in":"2024-01-06","nights":2,"selling_rate":150,"margin":20}]' http://localhost:8080/stats
    ```
5. **Stop:** Run `docker-compose down`.

## Architectural Decisions

*   **Project Layout:** The prokect uses the standard Go project layout (`cmd`, `internal`) for clear separation of concerns:
  
    *   `cmd/server`: Main application entry point and server setup.
    *   `internal/api`: Handles HTTP requests/responses and validation (API layer).
    *   `internal/booking`: Contains core domain logic, calculations, and the scheduling algorithm (Domain/Application layer).
    *   `internal/types`: Defines request/response DTOs.
    *   `internal/testutil`: Shared utilities for testing.
  
*   **Architecture Style:** The application leans towards a layered approach, separating API concerns from domain logic. While not a strict Hexagonal Architecture implementation (lacks explicit ports/adapters via interfaces), the separation allows for independent modification and testing of layers.
*   **Performance:** The core `/maximize` logic uses an efficient O(N log N) Dynamic Programming algorithm, suitable for handling large numbers of bookings (N) in a single request. The `/stats` endpoint is O(N). Memory usage is O(N). For extremely large single payloads (millions), API design changes (batching) might be needed over algorithm micro-optimization.

## Business Decisions

*   **Empty Requests:** An empty JSON array (`[]`) in the request body is considered valid input, resulting in a successful (200 OK) response with empty results.
*   **Validation:** Input validation checks are performed for mandatory fields: positive numbers (nights, rate, margin), and correct date formats.
*   **Error Handling:** Validation currently returns an error upon encountering the first issue found in the request list, providing immediate feedback but not a complete list of all problems. This decision was made taking into account that trying to return all the errors of a large input would be time-consuming while laying the same result: an error.
*   **Profit Calculation:** Total profit for a booking is calculated followint the formula `SellingRate * (Margin / 100.0)`. Profit per night (used in `/stats`) divides this result by the amount of `Nights`.

## Next Steps & Scalability

*   **Horizontal Scaling:** The application is stateless and can be scaled horizontally by running multiple Docker container instances behind a load balancer.
*   **Adding More Endpoints:** 
    *   New features (e.g., getting a specific booking, deleting) can be added by:
        1.  Defining new request/response types in `internal/types`.
        2.  Adding corresponding business logic functions in `internal/booking`.
        3.  Creating new handler functions in `internal/api`.
        4.  Registering the new routes in `cmd/server/main.go`.
*   **Refine Layout (Hexagonal):**
    *   Introduce interfaces (ports) for core services (e.g., `booking.SchedulerService`, `booking.StatsService`).
    *   Implement these interfaces in the `booking` package (adapters).
    *   Inject these interface dependencies into the `api` handlers. This further decouples layers and improves handler unit testability via mocking.
    *   Handlers could be split into separate files within `internal/api` (e.g., `maximize_handler.go`, `stats_handler.go`) as the number of endpoints grows.

## Areas for Improvement

*   **Validation Error Reporting:** Modify the validation logic (`validateAndMapBookings`) to collect *all* errors found in the request payload and return them in a single response, providing better feedback to clients.
*   **Handler Unit Testing:** Introduce interfaces for core services to allow for more isolated unit testing of the API handlers by mocking dependencies.
*   **Configuration:** Externalize configuration like the server port instead of hardcoding it.
*   **Health Check Endpoint:** Implement a dedicated `/health` endpoint for better monitoring and integration with orchestrators/load balancers.
*   **Logging & Monitoring:**
    *   Implement structured logging (e.g., using Go's standard `log/slog` package) throughout the application. Log key events like request start/end (with duration, status code, path), validation errors, internal errors, and panics with relevant context (like request IDs).
    *   Introduce application metrics (e.g., using a library like `prometheus/client_golang`). Track metrics such as request count, error rates (by status code/type), and request latency percentiles.
    *   Configure the application (or its deployment environment) to export logs and metrics to centralized systems (e.g., ELK stack, Loki, Prometheus, Grafana, Datadog) for effective monitoring, dashboarding, and alerting.