package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/tasqalent/tq-shared-go/errors"
	"github.com/tasqalent/tq-shared-go/logging"
	sharedmw "github.com/tasqalent/tq-shared-go/middleware"

	"github.com/tasqalent/tq-api-gateway/internal/config"
	gatewaymw "github.com/tasqalent/tq-api-gateway/internal/middleware"
	apiproxy "github.com/tasqalent/tq-api-gateway/internal/proxy"
)

func New(cfg config.Config) http.Handler {
	logging.Init(cfg.ServiceName, cfg.LogLevel)

	r := chi.NewRouter()

	r.Use(chimw.StripSlashes)

	r.Use(sharedmw.RequestID)
	r.Use(sharedmw.AccessLog)
	r.Use(sharedmw.SecurityHeaders)

	corsMW := sharedmw.NewCORS(sharedmw.CORSOptions{
		AllowedOrigins: cfg.CORSAllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type", sharedmw.HeaderRequestID},
		ExposedHeaders: []string{sharedmw.HeaderRequestID},
		AllowCredentials: cfg.CORSAllowCredentials,
		MaxAgeSeconds: 600,
	})
	r.Use(corsMW)

	jwtMW := gatewaymw.RequireBearerJWT(cfg.JWTSecret)
	rbacMW := gatewaymw.RequireRole(cfg.RequiredRole)

	protectedMW := func(next http.Handler) http.Handler {
		return jwtMW(rbacMW(next))
	}

	r.Use(gatewaymw.PublicPathSkipper(cfg.PublicPathPrefixes, protectedMW))

	r.NotFound(NotFoundJSON)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mustMount := func(pathPrefix, baseURL string) {
		if baseURL == "" {
			return
		}
		h, err := apiproxy.NewSingleHost(baseURL, cfg.ProxyTimeout)
		if err != nil {
			panic(pathPrefix + " upstream invalid: " + err.Error())
		}
		r.Mount(pathPrefix, http.StripPrefix(pathPrefix, h))
	}

	mustMount("/auth", cfg.AuthBaseURL)

	mustMount("/users", cfg.UsersBaseURL)
	mustMount("/gigs", cfg.GigBaseURL)
	mustMount("/chat", cfg.ChatBaseURL)
	mustMount("/orders", cfg.OrderBaseURL)
	mustMount("/reviews", cfg.ReviewBaseURL)

	mustMountWS := func(pathPrefix, baseURL string) {
		if baseURL == "" {
			return
		}
		h, err := apiproxy.NewWebSocketProxy(baseURL, cfg.WebSocketIdleTimeout)
		if err != nil {
			panic(pathPrefix + " ws upstream invalid: " + err.Error())
		}
		r.Mount(pathPrefix, http.StripPrefix(pathPrefix, h))
	}

	mustMountWS("/ws/chat", cfg.ChatWSBaseURL)
	mustMountWS("/ws/order", cfg.OrderWSBaseURL)
		
	return r
}

func NotFoundJSON(w http.ResponseWriter, _ *http.Request) {
	errors.WriteJSON(w, http.StatusNotFound, "not_found", "no route for this path", nil)
}