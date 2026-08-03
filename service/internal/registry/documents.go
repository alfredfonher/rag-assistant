package registry

import (
	"errors"

	"rag-assistant/service/internal/domain"
)

type documentRepo struct{ *Service }

func (s *Service) Documents() domain.DocumentRepo              { return documentRepo{s} }
func (r documentRepo) Get(id string) (*domain.Document, error) { return r.documents.Get(id) }
func (r documentRepo) List() ([]domain.Document, error)        { return r.documents.List() }
func (r documentRepo) ListByCollection(id string) ([]domain.Document, error) {
	return r.documents.ListByCollection(id)
}
func (r documentRepo) Create(document *domain.Document) error {
	return r.mutate(func() error {
		if _, err := r.collections.Get(document.CollectionID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrCollectionNotFound
			}
			return err
		}
		return r.documents.Create(document)
	})
}
func (r documentRepo) Update(document *domain.Document) error {
	return r.mutate(func() error {
		if _, err := r.collections.Get(document.CollectionID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrCollectionNotFound
			}
			return err
		}
		return r.documents.Update(document)
	})
}
func (r documentRepo) UpdateStatus(id, status string, chunks int, message string) error {
	return r.mutate(func() error { return r.documents.UpdateStatus(id, status, chunks, message) })
}
func (r documentRepo) Delete(id string) error {
	return r.mutate(func() error { return r.documents.Delete(id) })
}
