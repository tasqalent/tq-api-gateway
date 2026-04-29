package middleware

import (
	"net/http"
	"strings"
)

func IsPublicPath(path string, publicPrefixes []string) bool {
	for _, p := range publicPrefixes {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}

		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

func PublicPathSkipper(publicPrefixes []string, authMW func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		protected := authMW(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if IsPublicPath(r.URL.Path, publicPrefixes) {
				next.ServeHTTP(w, r)
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}