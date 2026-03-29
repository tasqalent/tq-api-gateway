package config

import (
	"time"

	sharedcfg "github.com/tasqalent/tq-shared-go/config"
)

type Config struct {
	ServiceName string
	HTTPAddr string
	LogLevel string
	AuthBaseURL string
	ProxyTimeout time.Duration
}

func Load() Config {
	return Config{
		ServiceName: sharedcfg.GetString("SERVICE_NAME", "tq-api-gateway"),
		HTTPAddr: sharedcfg.GetString("HTTP_ADDR", ":8080"),
		LogLevel: sharedcfg.GetString("LOG_LEVEL", "INFO"),
		AuthBaseURL: sharedcfg.GetString("AUTH_SERVICE_URL", "http://127.0.0.1:3001"),
		ProxyTimeout: sharedcfg.GetDuration("GATEWAY_PROXY_TIMEOUT", 30*time.Second),
	}
}

func FromEnv() Config {
	return Load()
}