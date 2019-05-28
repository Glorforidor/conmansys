package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Glorforidor/conmansys/insservice/storage"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("vim-go")
}

type postgres struct {
	db *sql.DB
}

func New(host, port, user, pass, dbname string) (*postgres, error) {
	connStr := fmt.Sprintf(
		"host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		host, port,
		user, pass,
		dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("could not open connection to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping database: %v", err)
	}

	return &postgres{
		db: db,
	}, nil
}

const (
	query = `
WITH RECURSIVE t(id) AS (
	SELECT dependee FROM conf_module_dependency 
	WHERE dependent=$1
	UNION
	SELECT d.dependee FROM conf_module_dependency AS d 
	INNER JOIN t ON t.id=d.dependent
)
SELECT conf_item.conf_item_value FROM t
JOIN conf_module ON t.id = conf_module.conf_module_id
JOIN conf_item_module ON conf_module.conf_module_id = conf_item_module.conf_module_id
JOIN conf_item ON conf_item_module.conf_item_id = conf_item.conf_item_id
UNION 
SELECT conf_item.conf_item_value FROM conf_module
JOIN conf_item_module ON conf_module.conf_module_id = conf_item_module.conf_module_id
JOIN conf_item ON conf_item_module.conf_item_id = conf_item.conf_item_id
WHERE conf_module.conf_module_id = $1
;`
)

// GetItems iterates over modules and find both modules' and their dependencies'
// items. Returns items and any error encountered.
func (p *postgres) GetItems(modules ...storage.Module) ([]*storage.Item, error) {
	// use the feature of a set to remove duplicates.
	set := make(map[string]*storage.Item)
	for _, m := range modules {
		rows, err := p.db.Query(query, m.ID)
		if err != nil {
			return nil, fmt.Errorf("could not execute query: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var it storage.Item
			err := rows.Scan(&it.Value)
			if err != nil {
				return nil, fmt.Errorf("could not scan data: %v", err)
			}

			// if needed could also say: set[it.Value+"@"+it.Version] to
			// distinguish on different versions.
			set[it.Value] = &it
		}

		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over rows: %v", err)
		}
	}

	var items []*storage.Item
	for _, v := range set {
		items = append(items, v)
	}

	return items, nil
}

const (
	itemsQuery = `
SELECT conf_item.conf_item_value FROM conf_module
JOIN conf_item_module ON conf_module.conf_module_id = conf_item_module.conf_module_id
JOIN conf_item ON conf_item_module.conf_item_id = conf_item.conf_item_id
WHERE conf_module.conf_module_id = $1
;
`

	moddepQuery = `
WITH RECURSIVE t(id) AS (
	SELECT dependee FROM conf_module_dependency 
	WHERE dependent=$1
	UNION
	SELECT d.dependee FROM conf_module_dependency AS d 
	INNER JOIN t ON t.id=d.dependent
)
SELECT * FROM t;`
)

// GetItemsAndModules finds items for the given modules and also any
// dependencies the modules might have. Returns slice of items and modules and
// an error if one has occured.
func (p *postgres) GetItemsAndModules(modules ...storage.Module) ([]*storage.Item, []*storage.Module, error) {
	set := make(map[string]interface{})
	for _, m := range modules {
		rows, err := p.db.Query(itemsQuery, m.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("could not execute query: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var it storage.Item
			err = rows.Scan(&it.Value)
			if err != nil {
				return nil, nil, fmt.Errorf("could not scan data: %v", err)
			}

			// if needed could also say: set[it.Value+"@"+it.Version] to
			// distinguish on different versions.
			set[it.Value] = &it
		}

		rows, err = p.db.Query(moddepQuery, m.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("could not execute query: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var mod storage.Module
			err = rows.Scan(&mod.ID)
			if err != nil {
				return nil, nil, fmt.Errorf("could not scan data: %v", err)
			}

			// if needed could also say: set[it.Value+"@"+it.Version] to
			// distinguish on different versions.
			set[string(mod.ID)] = &mod
		}

	}

	var items []*storage.Item
	var mods []*storage.Module
	for _, v := range set {
		switch vv := v.(type) {
		case *storage.Item:
			items = append(items, vv)
		case *storage.Module:
			mods = append(mods, vv)
		}
	}

	return items, mods, nil
}

func (p *postgres) Close() error {
	if err := p.db.Close(); err != nil {
		return fmt.Errorf("could not close database connection: %v", err)
	}
	return nil
}
