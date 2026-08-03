package registry

import (
	"sync"

	"rag-assistant/service/internal/domain"
)

type Service struct {
	agents      domain.AgentRepo
	collections domain.CollectionRepo
	documents   domain.DocumentRepo
	mutations   sync.Mutex
}

func New(a domain.AgentRepo, c domain.CollectionRepo, d domain.DocumentRepo) *Service {
	return &Service{agents: a, collections: c, documents: d}
}

func (s *Service) mutate(fn func() error) error {
	s.mutations.Lock()
	defer s.mutations.Unlock()
	return fn()
}
