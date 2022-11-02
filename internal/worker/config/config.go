package config

import "github.com/kackerx/crontab/internal/worker/options"

type Config struct {
	*options.Options
}

func NewConfig(opts *options.Options) (*Config, error) {
	return &Config{opts}, nil
}
