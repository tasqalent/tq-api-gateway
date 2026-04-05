package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	sharedhttp "github.com/tasqalent/tq-shared-go/httpclient"
	"github.com/tasqalent/tq-shared-go/middleware"
)

func NewSingleHost(baseURL string, timeout time.Duration) (http.Handler, error) {
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

	p.Transport = sharedhttp.New(timeout).Transport
	return p, nil
}