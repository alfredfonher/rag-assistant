package registry

import "rag-assistant/service/internal/domain"

type agentRepo struct{ *Service }

func (s *Service) Agents() domain.AgentRepo              { return agentRepo{s} }
func (r agentRepo) Get(id string) (*domain.Agent, error) { return r.agents.Get(id) }
func (r agentRepo) List() ([]domain.Agent, error)        { return r.agents.List() }
func (r agentRepo) Create(agent *domain.Agent) error {
	return r.mutate(func() error { return r.agents.Create(agent) })
}
func (r agentRepo) Update(agent *domain.Agent) error {
	return r.mutate(func() error { return r.agents.Update(agent) })
}
func (r agentRepo) Delete(id string) error {
	return r.mutate(func() error {
		collections, err := r.collections.ListByAgent(id)
		if err != nil {
			return err
		}
		if len(collections) != 0 {
			return domain.ErrAgentInUse
		}
		return r.agents.Delete(id)
	})
}
