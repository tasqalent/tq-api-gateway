package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/tasqalent/tq-shared-go/middleware"
)

func NewWebSocketProxy(baseURL string, idleTimeout time.Duration) (http.Handler, error) {
	target, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	p := httputil.NewSingleHostReverseProxy(target)

	orig := p.Rewrite
	p.Rewrite = func(pr *httputil.ProxyRequest) {
		if orig != nil {
			orig(pr)
		}

		if rid := pr.In.Header.Get(middleware.HeaderRequestID); rid != "" {
			pr.Out.Header.Set(middleware.HeaderRequestID, rid)
		}
	}

	p.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DisableCompression: false,
		MaxIdleConns: 100,
		IdleConnTimeout: idleTimeout,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return p, nil
}