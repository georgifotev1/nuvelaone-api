package middleware

import (
	"net/http"
	"strings"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
)

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				jsonutil.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			claims, err := auth.ParseToken(token, secret)
			if err != nil {
				jsonutil.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := auth.ContextWithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := auth.ClaimsFromContext(r.Context())
			if claims == nil {
				jsonutil.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			jsonutil.WriteError(w, http.StatusForbidden, "forbidden")
		})
	}
}

func extractToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
}
