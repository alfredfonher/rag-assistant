package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	ServiceName    string
	HTTPAddr       string
	DataDir        string
	IngestRoot     string
	LlamaServerURL string
}

func Default() Config {
	return Config{
		ServiceName:    "rag-assistant",
		HTTPAddr:       ":8080",
		DataDir:        ".rag-assistant",
		IngestRoot:     "docs",
		LlamaServerURL: "http://127.0.0.1:8090",
	}
}

func Load() (Config, error) {
	cfg := Default()

	if value := strings.TrimSpace(os.Getenv("RAG_SERVICE_NAME")); value != "" {
		cfg.ServiceName = value
	}

	if value := strings.TrimSpace(os.Getenv("RAG_HTTP_ADDR")); value != "" {
		cfg.HTTPAddr = value
	}

	if value := strings.TrimSpace(os.Getenv("RAG_DATA_DIR")); value != "" {
		cfg.DataDir = value
	}

	if value, ok := os.LookupEnv("RAG_INGEST_ROOT"); ok {
		cfg.IngestRoot = strings.TrimSpace(value)
	}

	if value := strings.TrimSpace(os.Getenv("RAG_LLAMA_SERVER_URL")); value != "" {
		cfg.LlamaServerURL = value
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.ServiceName) == "" {
		return fmt.Errorf("service name is required")
	}

	if strings.TrimSpace(c.HTTPAddr) == "" {
		return fmt.Errorf("http address is required")
	}

	if strings.TrimSpace(c.DataDir) == "" {
		return fmt.Errorf("data directory is required")
	}

	if strings.TrimSpace(c.IngestRoot) == "" {
		return fmt.Errorf("ingest root is required")
	}

	llamaURL, err := url.Parse(strings.TrimSpace(c.LlamaServerURL))
	if err != nil || (llamaURL.Scheme != "http" && llamaURL.Scheme != "https") || llamaURL.Host == "" {
		return fmt.Errorf("llama server URL must be an absolute http or https URL")
	}

	return nil
}
