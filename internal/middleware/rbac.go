package middleware

import (
	"net/http"

	"github.com/tasqalent/tq-shared-go/errors"
	"github.com/tasqalent/tq-shared-go/jwtutil"
)

func RequireRole(requiredRole string) func(http.Handler) http.Handler {
	if requiredRole == "" {
		return func (next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := jwtutil.ClaimsFromContext(r.Context())
			if !ok {
				errors.WriteJSON(w, http.StatusUnauthorized, "unauthorized", "missing token claims", nil)
				return
			}

			role, _ := claims["role"].(string)
			if role != requiredRole {
				errors.WriteJSON(w, http.StatusForbidden, "forbidden", "insufficient permissions", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}