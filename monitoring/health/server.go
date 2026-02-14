package health

import (
	"encoding/json"
	"net/http"
)

// LivenessHandler returns an HTTP handler for liveness checks.
// GET /health/live always returns 200 OK if the process is running.
func LivenessHandler(checker *Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := checker.Liveness()

		w.Header().Set("Content-Type", "application/health+json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(status)
	}
}

// ReadinessHandler returns an HTTP handler for readiness checks.
// GET /health/ready returns 200 OK if all checks pass, 503 Service Unavailable if any fail.
func ReadinessHandler(checker *Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		status := checker.Readiness(ctx)

		w.Header().Set("Content-Type", "application/health+json")

		if status.Status == "up" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(status)
	}
}
