package postgres

// TODO: rename some "want" variables since they don't reflect their intend.

import (
	"os"
	"testing"

	"github.com/Glorforidor/conmansys/confservice/storage"
)

var p *postgres

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
}

func TestMain(m *testing.M) {
	setup()

	os.Exit(m.Run())
}

func TestItem(t *testing.T) {
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

		moduleID := testCreateModule(t, tc.mValue, tc.mVersion)

		module := testGetModule(t, moduleID)

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

		itemModuleID := testCreateItemModule(t, itemID, moduleID)

		itemModule := testGetItemModule(t, itemModuleID)
		if itemModule.ItemID != itemID {
			t.Errorf("expected: %v, got: %v", itemID, itemModule.ItemID)
		}
		if itemModule.ModuleID != moduleID {
			t.Errorf("expected: %v, got: %v", moduleID, itemModule.ModuleID)
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

		row := testDeleteItemModule(t, itemModuleID)
		if row != 1 {
			t.Errorf("expected: %v, got: %v", 1, row)
		}

		row = testDeleteItem(t, itemID)
		if row != 1 {
			t.Errorf("expected: %v, got: %v", 1, row)
		}

		row = testDeleteModule(t, moduleID)
		if row != 1 {
			t.Errorf("expected: %v, got: %v", 1, row)
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

func testClose(t *testing.T) {
	err := p.Close()
	if err != nil {
		t.Fatalf("could not close database: %v", err)
	}
}
