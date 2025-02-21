package middleware

import (
	"net/http"
)

// LimitRequestBodyMiddleware restricts the size of the request body to the given `maxBytes` limit.
func LimitRequestBodyMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

			// Proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
