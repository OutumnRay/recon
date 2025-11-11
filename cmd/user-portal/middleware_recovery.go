package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// recoveryMiddleware catches panics and returns a 500 error
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error and stack trace
				fmt.Printf("PANIC: %v\n", err)
				fmt.Printf("Stack trace:\n%s\n", debug.Stack())

				// Return 500 error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Internal server error", "message": "Server panic occurred"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
