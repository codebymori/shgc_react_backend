package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"sentul-golf-be/config"
	"sentul-golf-be/models"
	"sentul-golf-be/utils"
)

type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware validates JWT token
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth check for OPTIONS requests (CORS preflight)
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.RespondUnauthorized(w, "Authorization header required")
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondUnauthorized(w, "Invalid authorization header format")
			return
		}

		token := parts[1]
		claims, err := utils.ValidateJWT(token, os.Getenv("JWT_SECRET"))
		if err != nil {
			utils.RespondUnauthorized(w, "Invalid or expired token")
			return
		}

		// Verify user still exists in database (not deleted)
		db := config.GetDB()
		var user models.User
		if err := db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
			utils.RespondUnauthorized(w, "User not found or has been deleted")
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole checks if user has required role
func RequireRole(role models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*utils.Claims)
			if !ok {
				utils.RespondUnauthorized(w, "Unauthorized")
				return
			}

			if claims.Role != string(role) && claims.Role != string(models.RoleAdmin) {
				utils.RespondForbidden(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is a helper for admin-only routes
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip for OPTIONS requests (already handled by CORS middleware)
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Apply role check
		RequireRole(models.RoleAdmin)(next).ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Allow localhost:3000 and any other origins (you can restrict this later)
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
		}
		
		// Check if origin is in allowed list
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}
		
		// Set CORS headers
		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			// Fallback to allow all origins (you can change this to deny if preferred)
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
