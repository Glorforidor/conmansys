package storage

type Service interface {
	GetItems(modules ...Module) ([]*Item, error)
}

type Item struct {
	ID      int64  `json:"id,omitempty"`
	Value   string `json:"value,omitempty"`
	Type    string `json:"type,omitempty"`
	Version string `json:"version,omitempty"`
}

type Module struct {
	ID      int64  `json:"id"`
	Value   string `json:"value,omitempty"`
	Version string `json:"version,omitempty"`
}
