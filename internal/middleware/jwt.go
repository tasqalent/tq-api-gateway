package middleware

import (
	"net/http"

	"github.com/tasqalent/tq-shared-go/errors"
	"github.com/tasqalent/tq-shared-go/jwtutil"
)

func RequireBearerJWT(secret string) func(http.Handler) http.Handler {
	if secret == "" {
		return func(next http.Handler) http.Handler { return next }
	}
	key := []byte(secret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := jwtutil.ParseBearerHS256(r.Header.Get("Authorization"), key)
			if err != nil {
				errors.WriteJSON(w, http.StatusUnauthorized, "unauthorized", "invalid or missing token", nil)
				return
			}
			ctx := jwtutil.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}