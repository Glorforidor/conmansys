package storage

type Service interface {
	GetItems(modules ...Module) (map[string]bool, error)
}

type Item struct {
	ID      int64  `json:"id"`
	Value   string `json:"value"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type Module struct {
	ID      int64  `json:"id"`
	Value   string `json:"value"`
	Version string `json:"version"`
}
