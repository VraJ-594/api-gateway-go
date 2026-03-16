package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"
	"golang.org/x/time/rate"
)
// create the wrapper to hold Limiter and last seen time for each IP address
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter manages individual token buckets for each IP address
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     rate.Limit // Tokens to add per second
	burst    int        // Maximum tokens the bucket can hold
}

// NewRateLimiter initializes the map and configuration
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    b,
	}

	go rl.cleanupVisitors() // Start the cleanup goroutine

	return rl
}

// Cleanup goroutine to remove old entries from the visitors map
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute) // Run cleanup every minute
		rl.mu.Lock()
		// Remove visitors that haven't been seen for more than 3 minutes
		for ip, vis := range rl.visitors {
			if time.Since(vis.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware wraps the handler with rate limiting logic
func (rl *RateLimiter) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		
		// 1. Extract the IP address (ignoring the port number)
		ip, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			log.Printf("Failed to parse IP: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// ALGORITHM CHALLENGE:
		// 2. Lock the mutex before touching the map: rl.mu.Lock()
		rl.mu.Lock()

		// 3. Check if the 'ip' exists in rl.visitors
		vis, exists := rl.visitors[ip]

		// 4. If it does not exist, create a new limiter and add it to the map:
		if !exists {

			limiter := rate.NewLimiter(rl.rate, rl.burst)
      rl.visitors[ip] = &visitor{limiter, time.Now()} 		
		} else {
			// 5. If it does exist, update the last seen time for this IP
			vis.lastSeen = time.Now()
		}
		// Get the limiter out of the struct to check it
		limiter := rl.visitors[ip].limiter

		// 6. Unlock the mutex: rl.mu.Unlock()  
		//    (Hint: NEVER forget to unlock, or your entire server will freeze forever!)
		rl.mu.Unlock()

		// 7. Check if the request is allowed
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// 8. If allowed, pass the request to the next handler
		next.ServeHTTP(w, req)
	})
}