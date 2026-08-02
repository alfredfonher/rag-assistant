package config

import "testing"

func TestLoad(t *testing.T) {
	tests := []struct {
		name   string
		setEnv func(t *testing.T)
		want   Config
	}{
		{
			name: "defaults",
			want: Default(),
		},
		{
			name: "environment overrides",
			setEnv: func(t *testing.T) {
				t.Setenv("RAG_SERVICE_NAME", "assistant")
				t.Setenv("RAG_HTTP_ADDR", "127.0.0.1:9000")
			},
			want: Config{ServiceName: "assistant", HTTPAddr: "127.0.0.1:9000", DataDir: ".rag-assistant", IngestRoot: "docs", LlamaServerURL: "http://127.0.0.1:8090"},
		},
		{
			name: "data directory override",
			setEnv: func(t *testing.T) {
				t.Setenv("RAG_DATA_DIR", "/tmp/rag-state")
			},
			want: Config{ServiceName: "rag-assistant", HTTPAddr: ":8080", DataDir: "/tmp/rag-state", IngestRoot: "docs", LlamaServerURL: "http://127.0.0.1:8090"},
		},
		{
			name: "ingest root override",
			setEnv: func(t *testing.T) {
				t.Setenv("RAG_INGEST_ROOT", "/srv/documents")
			},
			want: Config{ServiceName: "rag-assistant", HTTPAddr: ":8080", DataDir: ".rag-assistant", IngestRoot: "/srv/documents", LlamaServerURL: "http://127.0.0.1:8090"},
		},
		{
			name: "llama server URL override",
			setEnv: func(t *testing.T) {
				t.Setenv("RAG_LLAMA_SERVER_URL", "http://llama.internal:8090/")
			},
			want: Config{ServiceName: "rag-assistant", HTTPAddr: ":8080", DataDir: ".rag-assistant", IngestRoot: "docs", LlamaServerURL: "http://llama.internal:8090/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv != nil {
				tt.setEnv(t)
			}

			got, err := Load()
			if err != nil {
				t.Fatalf("Load() returned error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("Load() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestValidateRejectsEmptyIngestRoot(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.IngestRoot = "  "
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected empty ingest root to be rejected")
	}
}

func TestLoadRejectsEmptyIngestRootOverride(t *testing.T) {
	t.Setenv("RAG_INGEST_ROOT", "  ")
	if _, err := Load(); err == nil {
		t.Fatal("expected empty ingest root override to be rejected")
	}
}

func TestValidateRejectsInvalidLlamaServerURL(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"", "localhost:8090", "ftp://localhost/model", "http:///missing-host"} {
		value := value
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			cfg := Default()
			cfg.LlamaServerURL = value
			if err := cfg.Validate(); err == nil {
				t.Fatalf("expected %q to be rejected", value)
			}
		})
	}
}
