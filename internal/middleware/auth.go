package middleware

import (
	"log"
	"net/http"
)

// RequireAPIKey enforces that incoming requests provide a valid API key header
func RequireAPIKey(expectedKey string) func(http.Handler) http.Handler {
	
	// We return the actual middleware function
	return func(next http.Handler) http.Handler {
		
		// Which in turn returns the HTTP handler
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			
			// ALGORITHM CHALLENGE:
			// 1. Look for the header "X-API-Key" in req.Header
			// 2. Compare it to 'expectedKey'
			// 3. If they do NOT match, return http.StatusUnauthorized (401)
			//    and use log.Printf to log that an unauthorized attempt was blocked.
			// 4. If they DO match, call next.ServeHTTP(w, req)

			// Write your auth logic here...
			
			apiKey := req.Header.Get("X-API-Key")
			if apiKey != expectedKey {
				log.Printf("Unauthorized attempt blocked: %s", req.RemoteAddr)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}