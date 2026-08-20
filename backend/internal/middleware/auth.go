package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/abyankamal/sidak/backend/internal/service"
)

type contextKey string

const UserContextKey contextKey = "user_claims"

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := ""

			// 1. Check Bearer Authorization Header
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// 2. Fallback to Cookie "access_token"
			if tokenStr == "" {
				if cookie, err := r.Cookie("access_token"); err == nil && cookie.Value != "" {
					tokenStr = cookie.Value
				}
			}

			if tokenStr == "" {
				respondUnauthorized(w, "Sesi tidak sah atau token telah kedaluwarsa")
				return
			}

			claims, err := authService.ParseToken(tokenStr)
			if err != nil {
				respondUnauthorized(w, "Sesi tidak sah atau token telah kedaluwarsa")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*domain.AuthClaims)
			if !ok || claims == nil {
				respondUnauthorized(w, "Sesi tidak sah")
				return
			}

			isAllowed := false
			for _, role := range allowedRoles {
				if strings.EqualFold(claims.Role, role) {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(domain.ErrorResponse{
					Error: "Akses ditolak: peran pengguna tidak memiliki izin",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserClaims(r *http.Request) *domain.AuthClaims {
	if claims, ok := r.Context().Value(UserContextKey).(*domain.AuthClaims); ok {
		return claims
	}
	return nil
}

func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(domain.ErrorResponse{
		Error: message,
	})
}
