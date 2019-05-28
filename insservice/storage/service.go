package storage

import "fmt"

type Service interface {
	GetItems(modules ...Module) ([]*Item, error)
	GetItemsAndModules(modules ...Module) ([]*Item, []*Module, error)
}

type Item struct {
	ID      int64  `json:"id,omitempty"`
	Value   string `json:"value,omitempty"`
	Type    string `json:"type,omitempty"`
	Version string `json:"version,omitempty"`
}

func (i *Item) String() string {
	return fmt.Sprintf(
		"ID: %v, Value: %q, Type: %q, Version: %q",
		i.ID, i.Value, i.Type, i.Version,
	)
}

type Module struct {
	ID      int64  `json:"id"`
	Value   string `json:"value,omitempty"`
	Version string `json:"version,omitempty"`
}

func (m *Module) String() string {
	return fmt.Sprintf(
		"ID: %v, Value: %q, Version: %q",
		m.ID, m.Value, m.Version,
	)
}
