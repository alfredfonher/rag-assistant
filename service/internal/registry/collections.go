package registry

import (
	"errors"

	"rag-assistant/service/internal/domain"
)

type collectionRepo struct{ *Service }

func (s *Service) Collections() domain.CollectionRepo              { return collectionRepo{s} }
func (r collectionRepo) Get(id string) (*domain.Collection, error) { return r.collections.Get(id) }
func (r collectionRepo) List() ([]domain.Collection, error)        { return r.collections.List() }
func (r collectionRepo) ListByAgent(id string) ([]domain.Collection, error) {
	return r.collections.ListByAgent(id)
}
func (r collectionRepo) Create(collection *domain.Collection) error {
	return r.mutate(func() error {
		if err := r.requireAgent(collection.AgentID); err != nil {
			return err
		}
		return r.collections.Create(collection)
	})
}
func (r collectionRepo) Update(collection *domain.Collection) error {
	return r.mutate(func() error {
		if err := r.requireAgent(collection.AgentID); err != nil {
			return err
		}
		return r.collections.Update(collection)
	})
}
func (r collectionRepo) Delete(id string) error {
	return r.mutate(func() error {
		documents, err := r.documents.ListByCollection(id)
		if err != nil {
			return err
		}
		if len(documents) != 0 {
			return domain.ErrCollectionInUse
		}
		return r.collections.Delete(id)
	})
}

func (s *Service) requireAgent(id string) error {
	if _, err := s.agents.Get(id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrAgentNotFound
		}
		return err
	}
	return nil
}
