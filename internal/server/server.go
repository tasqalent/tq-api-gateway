package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/tasqalent/tq-shared-go/errors"
	"github.com/tasqalent/tq-shared-go/logging"
	"github.com/tasqalent/tq-shared-go/middleware"

	"github.com/tasqalent/tq-api-gateway/internal/config"
)

func New(cfg config.Config) http.Handler {
	logging.Init(cfg.ServiceName, cfg.LogLevel)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	if h, err := authProxyHandler(cfg); err == nil {
		mux.Handle("/auth/", h)
	} else {
		panic("AUTH_SERVICE_URL invalid: " + err.Error())
	}

	return middleware.RequestID(middleware.AccessLog(mux))
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
			pr.Out.Header.Set("X-Request-ID", rid)
		}
	}

	proxy.Transport = http.DefaultTransport

	return http.StripPrefix("/auth", proxy), nil
}

func NotFoundJSON(w http.ResponseWriter, _ *http.Request) {
	errors.WriteJSON(w, http.StatusNotFound, "not_found", "no route for this path", nil)
}