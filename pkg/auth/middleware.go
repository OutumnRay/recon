package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"Recontext.online/internal/models"
)

type contextKey string

const (
	// UserContextKey is the key for user claims in context
	UserContextKey contextKey = "user"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			// Extract token from Authorization header when provided
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || parts[0] != "Bearer" {
					respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
					return
				}
				token = parts[1]
			}

			// Fallback to token query parameter (e.g., for WebSocket connections)
			if token == "" {
				token = r.URL.Query().Get("token")
			}

			if token == "" {
				respondWithError(w, http.StatusUnauthorized, "Missing authorization token")
				return
			}

			// Verify token
			claims, err := jwtManager.VerifyToken(token)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole creates a middleware that checks for specific roles
func RequireRole(roles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*models.TokenClaims)
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "User not authenticated")
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondWithError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext retrieves user claims from context
func GetUserFromContext(ctx context.Context) (*models.TokenClaims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*models.TokenClaims)
	return claims, ok
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
