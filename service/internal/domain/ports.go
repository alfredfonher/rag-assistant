package domain

// AgentRepo is the port for agent persistence.
type AgentRepo interface {
	Get(id string) (*Agent, error)
	List() ([]Agent, error)
	Create(a *Agent) error
	Update(a *Agent) error
	Delete(id string) error
}

// CollectionRepo is the port for collection persistence.
type CollectionRepo interface {
	Get(id string) (*Collection, error)
	List() ([]Collection, error)
	ListByAgent(agentID string) ([]Collection, error)
	Create(c *Collection) error
	Update(c *Collection) error
	Delete(id string) error
}

// DocumentRepo is the port for document persistence.
type DocumentRepo interface {
	Get(id string) (*Document, error)
	List() ([]Document, error)
	ListByCollection(collectionID string) ([]Document, error)
	Create(d *Document) error
	Update(d *Document) error
	UpdateStatus(id string, status string, chunksCount int, errMsg string) error
	Delete(id string) error
}
