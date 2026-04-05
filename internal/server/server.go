package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/tasqalent/tq-shared-go/errors"
	"github.com/tasqalent/tq-shared-go/logging"
	"github.com/tasqalent/tq-shared-go/middleware"

	"github.com/tasqalent/tq-api-gateway/internal/config"
	apiproxy "github.com/tasqalent/tq-api-gateway/internal/proxy"
)

func New(cfg config.Config) http.Handler {
	logging.Init(cfg.ServiceName, cfg.LogLevel)

	r := chi.NewRouter()

	r.Use(chimw.StripSlashes)

	r.Use(middleware.RequestID)
	r.Use(middleware.AccessLog)
	r.Use(middleware.SecurityHeaders)

	corsMW := middleware.NewCORS(middleware.CORSOptions{
		AllowedOrigins: cfg.CORSAllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type", middleware.HeaderRequestID},
		ExposedHeaders: []string{middleware.HeaderRequestID},
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

	return r
}

func NotFoundJSON(w http.ResponseWriter, _ *http.Request) {
	errors.WriteJSON(w, http.StatusNotFound, "not_found", "no route for this path", nil)
}