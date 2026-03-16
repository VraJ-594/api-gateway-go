package routing

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic" // We need this for thread-safe counting
	"time"

	"my-gateway/internal/config" 
)
// BackendPool
type Backend struct {
	 URL  *url.URL
	 Proxy *httputil.ReverseProxy
	 Alive atomic.Bool
}

// SetAlive safely updates the health status
func (b *Backend) SetAlive(alive bool) {
	b.Alive.Store(alive)
}

// IsAlive safely reads the health status
func (b *Backend) IsAlive() bool {
	return b.Alive.Load()
}


// BackendPool holds multiple proxies for a single route
type BackendPool struct {
	backends []*Backend
	current uint64 // This is our atomic counter
}


type GatewayRouter struct {
	routes map[string]*BackendPool // Notice the map now points to a BackendPool
}

// NewRouter initializes the router with load-balanced pools
func NewRouter(cfg *config.Config) *GatewayRouter {
	router := &GatewayRouter{
		routes: make(map[string]*BackendPool),
	}

	for _, route := range cfg.Routes {
		pool := &BackendPool{
			backends: make([]*Backend, 0), 
			current:  0,
		}

		// Loop through every backend URL in the config array
		for _, backendURL := range route.Backends {

			parsedURL, err := url.Parse(backendURL)
			if err != nil {
				log.Fatalf("Invalid backend URL: %v", err)
			}

			proxy := &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					req.URL.Scheme = parsedURL.Scheme
					req.URL.Host = parsedURL.Host
					req.Host = parsedURL.Host
				},
			}

			backend := &Backend{
				URL:  parsedURL,
				Proxy: proxy,
			}
			backend.SetAlive(true) // Assume it's alive at startup
			backend.StartHealthCheck() // Start health checks for this backend

			pool.backends = append(pool.backends, backend)
		}

		// Save the completed pool into the router
		router.routes[route.Path] = pool
		log.Printf("Route registered: %s with %d backends", route.Path, len(pool.backends))
	}

	return router
}

func (r *GatewayRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	incomingPath := req.URL.Path

	var matchedPool *BackendPool
	longestMatchedPath := ""

	// 1. Find the longest matching path (your previous logic stays exactly the same)
	for registeredPath, pool := range r.routes {
		if strings.HasPrefix(incomingPath, registeredPath) {
			if len(registeredPath) > len(longestMatchedPath) {
				matchedPool = pool
				longestMatchedPath = registeredPath
			}
		}
	}

	// 2. If we found a match, do the Load Balancing math!
	if matchedPool != nil {
		
		// ALGORITHM CHALLENGE:
		// 1. Atomically add 1 to the matchedPool.current counter.
		//    Use: atomic.AddUint64(&matchedPool.current, 1)
		// 2. Figure out the index using the modulo operator (%).
		//    index := counter % uint64(len(matchedPool.proxies))
		// 3. Grab the proxy at that index: matchedPool.proxies[index]
		// 4. Call ServeHTTP on that proxy and return.
		
		// ALGORITHM CHALLENGE: Find an alive backend
		backendCount := uint64(len(matchedPool.backends))
		
		// Loop through the pool, starting from the next counter position
		for i := uint64(0); i < backendCount; i++ {
			counter := atomic.AddUint64(&matchedPool.current, 1)
			index := counter % backendCount
			backend := matchedPool.backends[index]

			// If we find an alive backend, route the request and exit
			if backend.IsAlive() {
				backend.Proxy.ServeHTTP(w, req)
				return
			}
		}

		http.Error(w, "Service Unavailable: All backends are down", http.StatusServiceUnavailable)
		return

	} else {
		http.Error(w, "Route not found", http.StatusNotFound)
	}
}


// StartHealthCheck pings the backend every 10 seconds
func (b *Backend) StartHealthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	
	go func() {
		for range ticker.C {
			// We use a 2-second timeout so a hanging server doesn't block us
			client := http.Client{Timeout: 2 * time.Second}
			
			// Ping the root URL of the backend
			resp, err := client.Get(b.URL.String())
			
			if err != nil || resp.StatusCode >= 500 {
				if b.IsAlive() {
					log.Printf("Backend went DOWN: %s", b.URL.String())
					b.SetAlive(false)
				}
			} else {
				if !b.IsAlive() {
					log.Printf("Backend recovered and is UP: %s", b.URL.String())
					b.SetAlive(true)
				}
			}
		}
	}()
}