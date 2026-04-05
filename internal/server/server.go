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

	r.NotFound(NotFoundJSON)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mustMount := func(router chi.Router, pathPrefix, baseURL string) {
		if baseURL == "" {
			return
		}
		h, err := apiproxy.NewSingleHost(baseURL, cfg.ProxyTimeout)
		if err != nil {
			panic(pathPrefix + " upstream invalid: " + err.Error())
		}
		router.Mount(pathPrefix, http.StripPrefix(pathPrefix, h))
	}

	mustMount(r, "/auth", cfg.AuthBaseURL)

	r.Group(func(r chi.Router) {
		r.Use(gatewaymw.RequireBearerJWT(cfg.JWTSecret))
		mustMount(r, "/users", cfg.UsersBaseURL)
		mustMount(r, "/gigs", cfg.GigBaseURL)
		mustMount(r, "/chat", cfg.ChatBaseURL)
		mustMount(r, "/orders", cfg.OrderBaseURL)
		mustMount(r, "/reviews", cfg.ReviewBaseURL)
	})
		
	return r
}

func NotFoundJSON(w http.ResponseWriter, _ *http.Request) {
	errors.WriteJSON(w, http.StatusNotFound, "not_found", "no route for this path", nil)
}