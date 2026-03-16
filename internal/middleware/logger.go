package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseRecorder intercepts the status code so we can log it
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before passing it to the real ResponseWriter
func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware wraps an existing handler with request logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		
		// 1. Create our custom recorder, defaulting to 200 OK
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// ALGORITHM CHALLENGE:
		// 2. Record the exact current time (start time)
		startTime := time.Now()

		// 3. Pass the request to the next handler in the chain.
		// CRITICAL: Pass 'recorder', not 'w'!
		next.ServeHTTP(recorder, req)

		// 4. Calculate how long the request took (current time - start time)
		duration := time.Since(startTime)

		// 5. Log the results in this format: 
		// [METHOD] /path | Status: 200 | Duration: 45ms
		log.Printf("[%s] %s | Status: %d | Duration: %v", req.Method, req.URL.Path, recorder.statusCode, duration)	
	})
}