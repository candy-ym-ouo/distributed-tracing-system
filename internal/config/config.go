package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr            string
	DataFile        string
	WebDir          string
	AuthToken       string
	Demo            bool
	Workers         int
	QueueSize       int
	ShutdownTimeout time.Duration
}

func Load(args []string) (Config, error) {
	cfg := Config{
		Addr:            env("TRACE_ADDR", ":8080"),
		DataFile:        env("TRACE_FILE", "data/spans.jsonl"),
		WebDir:          env("TRACE_WEB_DIR", "web"),
		AuthToken:       os.Getenv("TRACE_AUTH_TOKEN"),
		Workers:         envInt("TRACE_WORKERS", 4),
		QueueSize:       envInt("TRACE_QUEUE_SIZE", 4096),
		ShutdownTimeout: 5 * time.Second,
	}
	set := flag.NewFlagSet("trace-server", flag.ContinueOnError)
	set.StringVar(&cfg.Addr, "addr", cfg.Addr, "HTTP listen address")
	set.StringVar(&cfg.DataFile, "file", cfg.DataFile, "JSONL persistence file")
	set.StringVar(&cfg.WebDir, "web", cfg.WebDir, "static web directory")
	set.StringVar(&cfg.AuthToken, "auth-token", cfg.AuthToken, "optional upload token")
	set.BoolVar(&cfg.Demo, "demo", false, "generate demonstration traces")
	set.IntVar(&cfg.Workers, "workers", cfg.Workers, "collector worker count")
	set.IntVar(&cfg.QueueSize, "queue-size", cfg.QueueSize, "collector queue capacity")
	if err := set.Parse(args); err != nil {
		return Config{}, err
	}
	if cfg.Addr == "" {
		return Config{}, errors.New("address cannot be empty")
	}
	if cfg.Workers < 1 || cfg.Workers > 128 {
		return Config{}, errors.New("workers must be between 1 and 128")
	}
	if cfg.QueueSize < 1 {
		return Config{}, errors.New("queue size must be positive")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
