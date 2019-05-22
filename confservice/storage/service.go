package storage

type ItemService interface {
	GetItem(id int64) (*Item, error)
	GetItems() ([]*Item, error)
	CreateItem(value, iType, version string) (int64, error)
	DeleteItem(id int64) (int64, error)
}

type ModuleService interface {
	GetModule(id int64) (*Module, error)
	GetModules() ([]*Module, error)
	CreateModule(value, version string) (int64, error)
	DeleteModule(id int64) (int64, error)
}

type ItemModuleService interface {
	GetItemModule(id int64) (*ItemModule, error)
	GetItemModules() ([]*ItemModule, error)
	CreateItemModule(itemID, moduleID int64) (int64, error)
	DeleteItemModule(id int64) (int64, error)
}

type ModuleDependencyService interface {
	GetModuleDependencies() ([]*ModuleDependency, error)
	GetModuleDependenciesByDependentID(dependentID int64) ([]*ModuleDependency, error)
	GetModuleDependenciesByDependeeID(dependeeID int64) ([]*ModuleDependency, error)
	CreateModuleDependency(dependentID, dependeeID int64) error
	DeleteModuleDependency(dependentID, dependeeID int64) (int64, error)
	DeleteModuleDependencyByDependentID(id int64) (int64, error)
	DeleteModuleDependencyByDependeeID(id int64) (int64, error)
}

type Service interface {
	ItemService
	ModuleService
	ItemModuleService
	ModuleDependencyService
	Close() error
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

type ItemModule struct {
	ID       int64 `json:"id"`
	ItemID   int64 `json:"item_id"`
	ModuleID int64 `json:"module_id"`
}

type ModuleDependency struct {
	Dependent int64 `json:"dependent"`
	Dependee  int64 `json:"dependee"`
}
