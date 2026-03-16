package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"my-gateway/internal/config"
	"my-gateway/internal/routing"
	"my-gateway/internal/middleware"
)

// NewProxy takes a target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	// The Director modifies the incoming request before forwarding
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Step 1: Update the protocol (http vs https)
			req.URL.Scheme = url.Scheme
			
			// Step 2: Update the host to the target backend
			req.URL.Host = url.Host
			
			// Step 3: Ensure the Host header matches the target
			req.Host = url.Host 
			
			log.Printf("Proxying request to: %s", req.URL.String())
		},
	}

	return proxy, nil
}

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create Router , RateLimiter, and Auth Middleware
	router := routing.NewRouter(cfg)
  // ratelimiter with  1 request per second with a burst of 5
	rateLimiter := middleware.NewRateLimiter(1, 5)
	// auth
	authMiddleware := middleware.RequireAPIKey("my-super-secret-key")
  
	// 3. Chain the middlewares: 
	// RateLimit -> Auth -> Logging -> Router
	handler := rateLimiter.RateLimiterMiddleware(router)     // First line of defense
	handler = authMiddleware(handler)                   // Second line of defense
	handler = middleware.LoggingMiddleware(handler) // Record what passes

	port := ":8080"
	log.Printf("Gateway listening on port %s...", port)
	
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}