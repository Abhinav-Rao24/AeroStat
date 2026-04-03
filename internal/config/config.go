package config

import (
	"errors"
	"os"
	"time"
)

const (
	DefaultUnits = "metric"
	DefaultCacheTTL = 10 * time.Minute
	DefaultTimeout = 15 * time.Second
	envAPIKey = "OWM_API_KEY"
)

type Config struct {
	APIKey   string
	Units    string
	CacheTTL time.Duration
	Timeout  time.Duration
}

func Load() (*Config, error) {
	apiKey := os.Getenv(envAPIKey)
	if apiKey == "" {
		return nil, errors.New("OWM_API_KEY environment variable is not set")
	}

	units := os.Getenv("OWM_UNITS")
	if units == "" {
		units = DefaultUnits
	}

	return &Config{
		APIKey:   apiKey,
		Units:    units,
		CacheTTL: DefaultCacheTTL,
		Timeout:  DefaultTimeout,
	}, nil
}
