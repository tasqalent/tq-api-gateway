package config

import (
	"strings"
	"time"

	sharedcfg "github.com/tasqalent/tq-shared-go/config"
)

type Config struct {
	ServiceName string
	HTTPAddr string
	LogLevel string
	AuthBaseURL string
	ProxyTimeout time.Duration

	CORSAllowedOrigins []string
	CORSAllowCredentials bool
}

func Load() Config {
	return Config{
		ServiceName: sharedcfg.GetString("SERVICE_NAME", "tq-api-gateway"),
		HTTPAddr: sharedcfg.GetString("HTTP_ADDR", ":8080"),
		LogLevel: sharedcfg.GetString("LOG_LEVEL", "INFO"),
		AuthBaseURL: sharedcfg.GetString("AUTH_SERVICE_URL", "http://127.0.0.1:3001"),
		ProxyTimeout: sharedcfg.GetDuration("GATEWAY_PROXY_TIMEOUT", 30*time.Second),

		CORSAllowedOrigins: splitCSV(sharedcfg.GetString("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		CORSAllowCredentials: sharedcfg.GetString("CORS_ALLOW_CREDENTIALS", "false") == "true",
	}
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}