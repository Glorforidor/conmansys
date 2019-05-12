package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Glorforidor/conmansys/confservice/storage"
)

type postgres struct{ db *sql.DB }

func New(host, port, user, pass, dbname string) (storage.Service, error) {
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

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not ping database: %v", err)
	}

	return &postgres{
		db: db,
	}, nil
}

// TODO: make everything prepared. This is good practice though the driver might
// make everything prepared behind the curtain.

func (p *postgres) GetItem(id int64) (*storage.Item, error) {
	q := "SELECT * FROM conf_item WHERE conf_item_id = $1"

	i := &storage.Item{}

	err := p.db.QueryRow(q, id).Scan(
		&i.ID, &i.Value,
		&i.Type, &i.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("could not get item with id %v: %v", id, err)
	}

	return i, nil
}

func (p *postgres) GetItems() ([]*storage.Item, error) {
	q := "SELECT * FROM conf_item"

	rows, err := p.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("could not execute query: %v", err)
	}
	defer rows.Close()

	is := make([]*storage.Item, 0)

	for rows.Next() {
		i := &storage.Item{}
		err := rows.Scan(&i.ID, &i.Value, &i.Type, &i.Version)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %v", err)
		}
		is = append(is, i)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return is, nil
}

func (p *postgres) CreateItem(value, iType, version string) (int64, error) {
	q := `INSERT INTO conf_item
	(conf_item_value, conf_item_type, conf_item_version)
	VALUES ($1, $2, $3) RETURNING conf_item_id`

	// pq does not support LastInsertID. It is specified in the doc to use the
	// Query/QueryRow with the Postgres feauture of RETURNING.
	itemID := int64(0)
	err := p.db.QueryRow(q, value, iType, version).Scan(&itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			// this should properly not happend, unless postgres RETURNING
			// fails?
			return 0, nil
		}

		return 0, fmt.Errorf("could not insert data: %v", err)
	}

	return itemID, nil
}

func (p *postgres) DeleteItem(id int64) (int64, error) {
	q := "DELETE FROM conf_item WHERE conf_item_id = $1"

	rs, err := p.db.Exec(q, id)
	if err != nil {
		return 0, fmt.Errorf("could not delete item: %v", err)
	}

	count, err := rs.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("no rows were affected: %v", err)
	}

	return count, nil
}

func (p *postgres) GetModule(id int64) (*storage.Module, error) {
	q := "SELECT * FROM conf_module WHERE conf_module_id = $1"

	m := &storage.Module{}

	err := p.db.QueryRow(q, id).Scan(&m.ID, &m.Value, &m.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("could not get module with id %v: %v", id, err)
	}

	return m, nil
}

func (p *postgres) GetModules() ([]*storage.Module, error) {
	q := "SELECT * FROM conf_module"

	rows, err := p.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("could not execute q: %v", err)
	}
	defer rows.Close()

	ms := make([]*storage.Module, 0)

	for rows.Next() {
		m := &storage.Module{}
		err := rows.Scan(&m.ID, &m.Value, &m.Version)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %v", err)
		}
		ms = append(ms, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return ms, nil
}

func (p *postgres) CreateModule(value, version string) (int64, error) {
	q := `INSERT INTO conf_module (conf_module_value, conf_module_version)
	VALUES ($1, $2) RETURNING conf_module_id`

	moduleID := int64(0)
	err := p.db.QueryRow(q, value, version).Scan(&moduleID)
	if err != nil {
		if err == sql.ErrNoRows {
			// should properly not happend
			return 0, nil
		}

		return 0, fmt.Errorf("could not create module: %v", err)
	}

	return moduleID, nil
}

func (p *postgres) DeleteModule(id int64) (int64, error) {
	q := "DELETE FROM conf_module WHERE conf_module_id = $1"

	rs, err := p.db.Exec(q, id)
	if err != nil {
		return 0, fmt.Errorf("could not delete module: %v", err)
	}

	count, err := rs.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("no rows were affected: %v", err)
	}

	return count, nil
}

func (p *postgres) GetItemModule(id int64) (*storage.ItemModule, error) {
	q := "SELECT * FROM conf_item_module WHERE conf_item_module_id = $1"

	im := &storage.ItemModule{}

	err := p.db.QueryRow(q, id).Scan(&im.ID, &im.ItemID, &im.ModuleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("could not get itemModule with id %v: %v", id, err)
	}

	return im, nil
}

func (p *postgres) GetItemModules() ([]*storage.ItemModule, error) {
	q := "SELECT * FROM conf_item_module"

	rows, err := p.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("could not execute query: %v", err)
	}
	defer rows.Close()

	ims := make([]*storage.ItemModule, 0)

	for rows.Next() {
		im := &storage.ItemModule{}
		err := rows.Scan(&im.ID, &im.ItemID, &im.ModuleID)
		if err != nil {
			return nil, fmt.Errorf("could not get itemModules: %v", err)
		}

		ims = append(ims, im)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return ims, nil
}

func (p *postgres) CreateItemModule(itemID, moduleID int64) (int64, error) {
	q := `
	INSERT INTO conf_item_module (conf_item_id, conf_module_id) 
	VALUES ($1, $2)
	RETURNING conf_item_module_id`

	itemModuleID := int64(0)
	err := p.db.QueryRow(q, itemID, moduleID).Scan(&itemModuleID)
	if err != nil {
		if err == sql.ErrNoRows {
			// should properly not happend
			return 0, nil
		}

		return 0, fmt.Errorf("could not create ItemModule: %v", err)
	}

	return itemModuleID, nil
}

func (p *postgres) DeleteItemModule(id int64) (int64, error) {
	q := "DELETE FROM conf_item_module WHERE conf_item_module_id = $1"

	rs, err := p.db.Exec(q, id)
	if err != nil {
		return 0, fmt.Errorf("could not delete ItemModule: %v", err)
	}

	count, err := rs.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("no rows were affected: %v", err)
	}

	return count, nil
}

func (p *postgres) Close() error {
	return p.db.Close()
}
