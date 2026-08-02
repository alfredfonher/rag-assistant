package bootstrap

import (
	"path/filepath"

	"rag-assistant/service/internal/api"
	"rag-assistant/service/internal/config"
	"rag-assistant/service/internal/ingest"
	"rag-assistant/service/internal/llama"
	"rag-assistant/service/internal/persistence"
	"rag-assistant/service/internal/query"
	"rag-assistant/service/internal/readiness"
)

type App struct {
	Config config.Config
	Server *api.Server
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	llamaClient := llama.NewClient(cfg.LlamaServerURL)
	embedder := llama.NewEmbedder(llamaClient)
	persistentRetriever, err := query.NewPersistentMemoryRetrieverWithProvider(filepath.Join(cfg.DataDir, "index.json"), embedder)
	if err != nil {
		return nil, err
	}
	conversationStore, err := query.NewConversationStore(filepath.Join(cfg.DataDir, "conversations.json"))
	if err != nil {
		return nil, err
	}

	// Domain repos
	agentRepo, err := persistence.NewFileAgentRepo(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	collectionRepo, err := persistence.NewFileCollectionRepo(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	documentRepo, err := persistence.NewFileDocumentRepo(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	ingestService, err := ingest.New(persistentRetriever, cfg.IngestRoot)
	if err != nil {
		return nil, err
	}

	server := api.NewServerWithDeps(cfg.ServiceName, api.ServerDeps{
		Readiness:     readiness.LlamaReadiness{Client: llamaClient},
		Ingest:        ingestService,
		Query:         query.New(persistentRetriever, llama.NewComposer(llamaClient), conversationStore),
		Agents:        agentRepo,
		Collections:   collectionRepo,
		Documents:     documentRepo,
		Conversations: conversationStore,
	})

	return &App{
		Config: cfg,
		Server: server,
	}, nil
}
