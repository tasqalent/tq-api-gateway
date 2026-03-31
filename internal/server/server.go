package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/tasqalent/tq-shared-go/errors"
	sharedhttp "github.com/tasqalent/tq-shared-go/httpclient"
	"github.com/tasqalent/tq-shared-go/logging"
	"github.com/tasqalent/tq-shared-go/middleware"

	"github.com/tasqalent/tq-api-gateway/internal/config"
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

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	authProxy, err := authProxyHandler(cfg)
	if err != nil {
		panic("AUTH_SERVICE_URL invalid: " + err.Error())
	}
	r.Mount("/auth", http.StripPrefix("/auth", authProxy))

	return r
}

func authProxyHandler(cfg config.Config) (http.Handler, error) {
	target, err := url.Parse(cfg.AuthBaseURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalRewrite := proxy.Rewrite
	proxy.Rewrite = func(pr *httputil.ProxyRequest) {
		if originalRewrite != nil {
			originalRewrite(pr)
		}
		if rid := pr.In.Header.Get(middleware.HeaderRequestID); rid != "" {
			pr.Out.Header.Set(middleware.HeaderRequestID, rid)
		}
	}

	proxy.Transport = sharedhttp.New(cfg.ProxyTimeout).Transport

	return proxy, nil
}

func NotFoundJSON(w http.ResponseWriter, _ *http.Request) {
	errors.WriteJSON(w, http.StatusNotFound, "not_found", "no route for this path", nil)
}