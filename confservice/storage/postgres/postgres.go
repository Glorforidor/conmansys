package postgres

import (
	"database/sql"
	"fmt"

	// load postgres driver
	_ "github.com/lib/pq"

	"github.com/Glorforidor/conmansys/confservice/storage"
)

type postgres struct{ db *sql.DB }

// New return new initialised storage.Service. If there is an error it will be a
// connection error.
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

// GetItem finds the item with the given id in the database and returns it. If
// an error occurs it returns nil and the error.
func (p *postgres) GetItem(id int64) (*storage.Item, error) {
	q := "SELECT * FROM conf_item WHERE conf_item_id = $1"

	i := &storage.Item{}

	err := p.db.QueryRow(q, id).Scan(
		&i.ID, &i.Value,
		&i.Type, &i.Version,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("could not get item with id %v: %v", id, err)
	}

	return i, nil
}

// GetItems finds every item in the database and returns a slice of items. If an
// error occurs it returns nil slice and the error.
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

// TODO: maybe add Stringer to structs so createtype takes a stringer instead of
// string so input is more reliable?

func create(db *sql.DB, query string, createType string, args ...interface{}) (int64, error) {
	i := int64(0)
	err := db.QueryRow(query, args...).Scan(&i)
	if err != nil {
		if err == sql.ErrNoRows {
			// should properly not happend
			return 0, nil
		}

		return 0, fmt.Errorf("could not create %v: %v", createType, err)
	}

	return i, nil
}

// CreateItem inserts a new row into the database and return the id of the new
// created row. If an error occurs the returned id is 0 and the insertion error.
func (p *postgres) CreateItem(value, iType, version string) (int64, error) {
	q := `INSERT INTO conf_item
	(conf_item_value, conf_item_type, conf_item_version)
	VALUES ($1, $2, $3) RETURNING conf_item_id`

	return create(p.db, q, "Item", value, iType, version)
}

func delete(db *sql.DB, query string, deleteType string, args ...interface{}) (int64, error) {
	rs, err := db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("could not delete %v: %v", deleteType, err)
	}

	count, err := rs.RowsAffected()
	// TODO: should this be an error if no rows were affected?
	if err != nil {
		return 0, fmt.Errorf("no rows were affected: %v", err)
	}

	return count, nil
}

// DeleteItem deletes the item with the given id in the database. It returns the
// affected rows. If no rows were affected it is considered as an error.
func (p *postgres) DeleteItem(id int64) (int64, error) {
	q := "DELETE FROM conf_item WHERE conf_item_id = $1"

	return delete(p.db, q, "Item", id)
}

// GetModule finds the module with the given id in the database and returns it.
// If an error occurs it returns nil and the error.
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

// GetModules find modules in the database and returns slice of modules. If an
// error occurs it return nil slice and the error.
func (p *postgres) GetModules() ([]*storage.Module, error) {
	q := "SELECT * FROM conf_module"

	rows, err := p.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("could not execute query: %v", err)
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

// CreateModule inserts a module with the given values into the database and
// returns the newly inserted modules id. If an error occurs the id will be 0
// and the caused error.
func (p *postgres) CreateModule(value, version string) (int64, error) {
	q := `INSERT INTO conf_module (conf_module_value, conf_module_version)
	VALUES ($1, $2) RETURNING conf_module_id`

	return create(p.db, q, "Module", value, version)
}

// DeleteModule deletes the module with the given id in the database and returns
// the rows affected. If 0 rows are affected it is treated as an error.
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

// GetItemModule finds the item module in the database and returns the it. If
// an error occurs it returns nil and the error.
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

// GetItemModules find the item modules in the database and returns slice of
// item modules. If and error occurs it returns nil slice and the error.
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

// CreateItemModule inserts a item module with the given values and returns the
// newly inserted item module's id. If an error occurs it returns 0 and the
// error.
func (p *postgres) CreateItemModule(itemID, moduleID int64) (int64, error) {
	q := `
	INSERT INTO conf_item_module (conf_item_id, conf_module_id) 
	VALUES ($1, $2)
	RETURNING conf_item_module_id`

	return create(p.db, q, "ItemModule", itemID, moduleID)
}

// DeleteItemModule deletes the item module with the given id and returns the
// rows affected. If 0 rows are affected it is treated as an error.
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

func modDep(db *sql.DB, query string, args ...interface{}) ([]*storage.ModuleDependency, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not execute query: %v", err)
	}
	defer rows.Close()

	mds := make([]*storage.ModuleDependency, 0)
	for rows.Next() {
		md := &storage.ModuleDependency{}
		err := rows.Scan(&md.Dependent, &md.Dependee)
		if err != nil {
			return nil, fmt.Errorf("could not get module dependencies: %v", err)
		}

		mds = append(mds, md)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return mds, nil
}

func (p *postgres) GetModuleDependencies() ([]*storage.ModuleDependency, error) {
	q := "SELECT * FROM conf_module_dependency"

	return modDep(p.db, q)
}

// GetModuleDependenciesByDependentID finds module dependency by dependent id
// and returns slice of module dependencies. If an error occurs it returns nil
// slice and the error.
func (p *postgres) GetModuleDependenciesByDependentID(dependentID int64) ([]*storage.ModuleDependency, error) {
	q := "SELECT * FROM conf_module_dependency WHERE dependent = $1"

	return modDep(p.db, q, dependentID)
}

// GetModuleDependenciesByDependentID finds module dependency by dependee id
// and returns slice of module dependencies. If an error occurs it returns nil
// slice and the error.
func (p *postgres) GetModuleDependenciesByDependeeID(dependeeID int64) ([]*storage.ModuleDependency, error) {
	q := "SELECT * FROM conf_module_dependency WHERE dependee = $1"

	return modDep(p.db, q, dependeeID)
}

// CreateModuleDependency inserts a module dependency with given dependent and
// dependee id. If an error occurs it could not create the module dependency.
func (p *postgres) CreateModuleDependency(dependentID int64, dependeeID int64) error {
	q := "INSERT INTO conf_module_dependency VALUES ($1, $2)"

	_, err := create(p.db, q, "ModuleDependency", dependentID, dependeeID)

	return err
}

// DeleteModuleDependency deletes the module dependency with the given dependent
// and dependee id and returns rows affected. If 0 rows were affected it is
// treated as an error.
func (p *postgres) DeleteModuleDependency(dependentID, dependeeID int64) (int64, error) {
	q := "DELETE FROM conf_module_dependency WHERE dependent = $1 AND dependee = $2"

	return delete(p.db, q, "ModuleDependency", dependentID, dependeeID)
}

// DeleteModuleDependencyByDependentID deletes the module dependency with the
// given dependent id and returns rows affected. If 0 rows were affected it is
// treated as an error.
func (p *postgres) DeleteModuleDependencyByDependentID(id int64) (int64, error) {
	q := "DELETE FROM conf_module_dependency WHERE dependent = $1"

	return delete(p.db, q, "ModuleDependency", id)
}

// DeleteModuleDependencyByDependeeID deletes the module dependency with the
// given dependee id and returns rows affected. If 0 rows were affected it is
// treated as an error.
func (p *postgres) DeleteModuleDependencyByDependeeID(id int64) (int64, error) {
	q := "DELETE FROM conf_module_dependency WHERE dependee = $1"

	return delete(p.db, q, "ModuleDependency", id)
}

// Close closes the database connection.
func (p *postgres) Close() error {
	return p.db.Close()
}
