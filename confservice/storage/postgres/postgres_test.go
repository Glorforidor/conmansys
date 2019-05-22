package postgres

// TODO: rename some "want" variables since they don't reflect their intend.

import (
	"os"
	"testing"

	"github.com/Glorforidor/conmansys/confservice/storage"
)

var p *postgres

const dataschema = `
DROP TABLE IF EXISTS conf_item_module;
DROP TABLE IF EXISTS conf_item;
DROP TABLE IF EXISTS conf_module;
DROP TABLE IF EXISTS conf_module_dependency;

CREATE TABLE conf_item(
	conf_item_id SERIAL PRIMARY KEY,
	conf_item_value TEXT NOT NULL,
	conf_item_type TEXT NOT NULL,
	conf_item_version TEXT NOT NULL
);

CREATE TABLE conf_module(
	conf_module_id SERIAL PRIMARY KEY,
	conf_module_value TEXT NOT NULL,
	conf_module_version TEXT NOT NULL
);

CREATE TABLE conf_module_dependency(
	dependent int,
	dependee int,
	PRIMARY KEY (dependent, dependee)
);

CREATE TABLE conf_item_module(
	conf_item_module_id SERIAL PRIMARY KEY,
	conf_item_id INTEGER,
	conf_module_id INTEGER,
	FOREIGN KEY (conf_item_id) REFERENCES conf_item(conf_item_id) ON DELETE CASCADE,
	FOREIGN KEY (conf_module_id) REFERENCES conf_module(conf_module_id) ON DELETE CASCADE
);
`

// there must be a better way?
func setup() {
	host := "172.17.0.2"
	port := "5432"
	user := "postgres"
	pass := "secret"
	dbname := "conmansys"

	pp, err := New(host, port, user, pass, dbname)
	if err != nil {
		panic(err)
	}

	p = pp.(*postgres)

	p.db.Exec("DROP DATABASE IF EXISTS conmansys")
	p.db.Exec("CREATE DATABASE conmansys")

	_, err = p.db.Exec(dataschema)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	setup()

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	tt := map[string]struct {
		host   string
		port   string
		user   string
		pass   string
		dbname string
		err    bool
	}{
		"good connection": {
			host:   "172.17.0.2", // important a test database is running
			port:   "5432",
			user:   "postgres",
			pass:   "secret",
			dbname: "conmansys",
		},
		"bad connection": {
			host:   "172.17.0.2",
			port:   "5000",
			user:   "postgres",
			pass:   "secret",
			dbname: "conmansys",
			err:    true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			postgres, err := New(tc.host, tc.port, tc.user, tc.pass, tc.dbname)
			if err != nil {
				if !tc.err {
					t.Fatalf("expected error, got: %v", err)
				}
				return
			}

			if postgres == nil {
				t.Fatalf("expected initialised postgres, got: %v", postgres)
			}

			postgres.Close()
		})
	}
}

// integration test! seems easier for database testing
func TestEverything(t *testing.T) {
	tt := []struct {
		iValue   string
		iType    string
		iVersion string
		mValue   string
		mVersion string
	}{
		{
			iValue:   "posty",
			iType:    "database",
			iVersion: "0.0.1",
			mValue:   "posty_mod",
			mVersion: "0.0.1",
		},
	}

	for _, tc := range tt {
		itemID := testCreateItem(t, tc.iValue, tc.iType, tc.iVersion)

		item := testGetItem(t, itemID)

		if item.Value != tc.iValue {
			t.Errorf("expected: %v, got: %v", tc.iValue, item.Value)
		}
		if item.Type != tc.iType {
			t.Errorf("expected: %v, got: %v", tc.iValue, item.Value)
		}
		if item.Version != tc.iVersion {
			t.Errorf("expected: %v, got: %v", tc.iValue, item.Value)
		}

		items := testGetItems(t)

		found := false
		for _, i := range items {
			if *i == *item {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected: %v, was not in list: %v", item, items)
		}

		moduleID1 := testCreateModule(t, tc.mValue, tc.mVersion)
		moduleID2 := testCreateModule(t, tc.mValue, tc.mVersion)
		moduleID3 := testCreateModule(t, tc.mValue, tc.mVersion)
		moduleID4 := testCreateModule(t, tc.mValue, tc.mVersion)
		moduleID5 := testCreateModule(t, tc.mValue, tc.mVersion)
		moduleID6 := testCreateModule(t, tc.mValue, tc.mVersion)

		module := testGetModule(t, moduleID1)

		if module.Value != tc.mValue {
			t.Errorf("expected: %v, got: %v", tc.mValue, module.Value)
		}
		if module.Version != tc.mVersion {
			t.Errorf("expected: %v, got: %v", tc.mVersion, module.Version)
		}

		modules := testGetModules(t)

		found = false
		for _, m := range modules {
			if *m == *module {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected: %v, was not in list: %v", module, modules)
		}

		itemModuleID := testCreateItemModule(t, itemID, moduleID1)

		itemModule := testGetItemModule(t, itemModuleID)
		if itemModule.ItemID != itemID {
			t.Errorf("expected: %v, got: %v", itemID, itemModule.ItemID)
		}
		if itemModule.ModuleID != moduleID1 {
			t.Errorf("expected: %v, got: %v", moduleID1, itemModule.ModuleID)
		}

		itemModules := testGetItemModules(t)

		found = false
		for _, im := range itemModules {
			if *im == *itemModule {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected: %v, was not in list: %v", itemModule, itemModules)
		}

		testCreateModuleDependency(t, moduleID1, moduleID2)
		testCreateModuleDependency(t, moduleID3, moduleID4)
		testCreateModuleDependency(t, moduleID5, moduleID6)

		// perhaps a better way to test this.
		moddeps1 := testGetModuleDependecies(t)

		for _, moddep := range moddeps1 {
			if moddep.Dependent != moduleID1 && moddep.Dependent != moduleID3 && moddep.Dependent != moduleID5 {
				t.Errorf("expected: %v, got: %v", moduleID1, moddep.Dependent)
			}
			if moddep.Dependee != moduleID2 && moddep.Dependee != moduleID4 && moddep.Dependee != moduleID6 {
				t.Errorf("expected: %v, got: %v", moduleID2, moddep.Dependent)
			}
		}

		moddeps2 := testGetModuleDependenciesByDependentID(t, moduleID1)
		moddeps3 := testGetModuleDependenciesByDependeeID(t, moduleID2)

		for i, moddep := range moddeps2 {
			if *moddep != *moddeps3[i] {
				t.Errorf("expected to be equal, got: (%v, %v)", *moddep, *moddeps3[i])
			}
		}

		row := testDeleteItemModule(t, itemModuleID)
		if row != 1 {
			t.Errorf("expected: %v, got: %v", 1, row)
		}

		row = testDeleteItem(t, itemID)
		if row != 1 {
			t.Errorf("expected: %v, got: %v", 1, row)
		}

		row = testDeleteModule(t, moduleID1)
		if row != 1 {
			t.Errorf("expected: %v, got: %v", 1, row)
		}

		row = testDeleteDependency(t, moduleID1, moduleID2)
		if row != 1 {
			t.Errorf("expected: 1, got: %v", row)
		}

		row = testDeleteModuleDependecyByDependentID(t, moduleID3)
		if row != 1 {
			t.Errorf("expected: 1, got: %v", row)
		}
		row = testDeleteModuleDependecyByDependeeID(t, moduleID6)
		if row != 1 {
			t.Errorf("expected: 1, got: %v", row)
		}
	}

	testClose(t)
}

func testGetItem(t *testing.T, id int64) *storage.Item {
	item, err := p.GetItem(id)
	if err != nil {
		t.Fatalf("could not get item with id:%v: %v", id, err)
	}
	return item
}

func testGetItems(t *testing.T) []*storage.Item {
	items, err := p.GetItems()
	if err != nil {
		t.Fatalf("could not get items: %v", err)
	}
	return items
}

func testCreateItem(t *testing.T, value, iType, version string) int64 {
	id, err := p.CreateItem(value, iType, version)
	if err != nil {
		t.Fatalf(
			"could not create item with values (%v, %v, %v): %v",
			value, iType, version, err,
		)
	}
	return id
}

func testDeleteItem(t *testing.T, id int64) int64 {
	row, err := p.DeleteItem(id)
	if err != nil {
		t.Fatalf("could not delete item with id %v: %v", id, err)
	}
	return row
}

func testGetModule(t *testing.T, id int64) *storage.Module {
	m, err := p.GetModule(id)
	if err != nil {
		t.Fatalf("could not get module with id %v: %v", id, err)
	}
	return m
}

func testGetModules(t *testing.T) []*storage.Module {
	mm, err := p.GetModules()
	if err != nil {
		t.Fatalf("could not get modules: %v", err)
	}
	return mm
}

func testCreateModule(t *testing.T, value, version string) int64 {
	i, err := p.CreateModule(value, version)
	if err != nil {
		t.Fatalf(
			"could not create module with values (%v, %v): %v",
			value,
			version,
			err,
		)
	}
	return i
}

func testDeleteModule(t *testing.T, id int64) int64 {
	row, err := p.DeleteModule(id)
	if err != nil {
		t.Fatalf("could not delete module with id %v: %v", id, err)
	}
	return row
}

func testGetItemModule(t *testing.T, id int64) *storage.ItemModule {
	im, err := p.GetItemModule(id)
	if err != nil {
		t.Fatalf("could not get item_module with id %v: %v", id, err)
	}
	return im
}

func testGetItemModules(t *testing.T) []*storage.ItemModule {
	ims, err := p.GetItemModules()
	if err != nil {
		t.Fatalf("could not get item_modules: %v", err)
	}
	return ims
}

func testCreateItemModule(t *testing.T, itemID, moduleID int64) int64 {
	id, err := p.CreateItemModule(itemID, moduleID)
	if err != nil {
		t.Fatalf(
			"could not create item_module with values (%v, %v): %v",
			itemID,
			moduleID,
			err,
		)
	}
	return id
}

func testDeleteItemModule(t *testing.T, id int64) int64 {
	row, err := p.DeleteItemModule(id)
	if err != nil {
		t.Errorf("could not delete item_module with id %v: %v", id, err)
	}
	return row
}

func testGetModuleDependecies(t *testing.T) []*storage.ModuleDependency {
	m, err := p.GetModuleDependencies()
	if err != nil {
		t.Fatalf("could not get module_dependencies: %v", err)
	}

	return m
}

func testGetModuleDependenciesByDependentID(t *testing.T, id int64) []*storage.ModuleDependency {
	m, err := p.GetModuleDependenciesByDependentID(id)
	if err != nil {
		t.Fatalf("could not get module_dependencies: %v", err)
	}

	return m
}

func testGetModuleDependenciesByDependeeID(t *testing.T, id int64) []*storage.ModuleDependency {
	m, err := p.GetModuleDependenciesByDependeeID(id)
	if err != nil {
		t.Fatalf("could not get module_dependencies: %v", err)
	}

	return m
}

func testCreateModuleDependency(t *testing.T, depedentID, dependeeID int64) {
	err := p.CreateModuleDependency(depedentID, dependeeID)
	if err != nil {
		t.Fatalf("could not create module_dependency: %v", err)
	}
}

func testDeleteDependency(t *testing.T, dependentID, dependeeID int64) int64 {
	i, err := p.DeleteModuleDependency(dependentID, dependeeID)
	if err != nil {
		t.Fatalf("could not delete module_dependency: %v", err)
	}

	return i
}

func testDeleteModuleDependecyByDependentID(t *testing.T, id int64) int64 {
	i, err := p.DeleteModuleDependencyByDependentID(id)
	if err != nil {
		t.Fatalf("could not delete module_dependency: %v", err)
	}

	return i
}

func testDeleteModuleDependecyByDependeeID(t *testing.T, id int64) int64 {
	i, err := p.DeleteModuleDependencyByDependeeID(id)
	if err != nil {
		t.Fatalf("could not delete module_dependency: %v", err)
	}

	return i
}

func testClose(t *testing.T) {
	err := p.Close()
	if err != nil {
		t.Fatalf("could not close database: %v", err)
	}
}
