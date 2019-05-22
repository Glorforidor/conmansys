package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Glorforidor/conmansys/confservice/storage"
)

type dbmock struct {
	items        []*storage.Item
	modules      []*storage.Module
	itemModules  []*storage.ItemModule
	dependencies []*storage.ModuleDependency
	closed       bool
}

func (d *dbmock) GetItem(id int64) (*storage.Item, error) {
	if d.closed {
		return nil, errors.New("")
	}
	return d.items[id-1], nil
}

func (d *dbmock) GetItems() ([]*storage.Item, error) {
	if d.closed {
		return nil, errors.New("")
	}
	return d.items, nil
}

func (d *dbmock) CreateItem(value string, iType string, version string) (int64, error) {
	if d.closed {
		return 0, errors.New("")
	}
	return 1, nil
}

func (d *dbmock) DeleteItem(id int64) (int64, error) {
	if d.closed {
		return 0, errors.New("")
	}
	return 1, nil
}

func (d *dbmock) GetModule(id int64) (*storage.Module, error) {
	if d.closed {
		return nil, errors.New("")
	}
	return d.modules[id-1], nil
}

func (d *dbmock) GetModules() ([]*storage.Module, error) {
	if d.closed {
		return nil, errors.New("")
	}
	return d.modules, nil
}

func (d *dbmock) CreateModule(value string, version string) (int64, error) {
	if d.closed {
		return 0, errors.New("")
	}
	return 1, nil
}

func (d *dbmock) DeleteModule(id int64) (int64, error) {
	if d.closed {
		return 0, errors.New("")
	}
	return 1, nil
}

func (d *dbmock) GetItemModule(id int64) (*storage.ItemModule, error) {
	if d.closed {
		return nil, errors.New("")
	}
	return d.itemModules[id-1], nil
}

func (d *dbmock) GetItemModules() ([]*storage.ItemModule, error) {
	if d.closed {
		return nil, errors.New("")
	}
	return d.itemModules, nil
}

func (d *dbmock) CreateItemModule(itemID int64, moduleID int64) (int64, error) {
	if d.closed {
		return 0, errors.New("")
	}
	return 1, nil
}

func (d *dbmock) DeleteItemModule(id int64) (int64, error) {
	if d.closed {
		return 0, errors.New("")
	}
	return 1, nil
}

func (d *dbmock) GetModuleDependencies() ([]*storage.ModuleDependency, error) {
	if d.closed {
		return nil, errors.New("")
	}

	return d.dependencies, nil
}

func (d *dbmock) GetModuleDependenciesByDependentID(dependentID int64) ([]*storage.ModuleDependency, error) {
	if d.closed {
		return nil, errors.New("")
	}

	deps := make([]*storage.ModuleDependency, len(d.dependencies))
	for _, dep := range d.dependencies {
		if dep.Dependent == dependentID {
			deps = append(deps, dep)
		}

	}

	return deps, nil
}

func (d *dbmock) GetModuleDependenciesByDependeeID(dependeeID int64) ([]*storage.ModuleDependency, error) {
	if d.closed {
		return nil, errors.New("")
	}

	deps := make([]*storage.ModuleDependency, len(d.dependencies))
	for _, dep := range d.dependencies {
		if dep.Dependee == dependeeID {
			deps = append(deps, dep)
		}

	}

	return deps, nil
}

func (d *dbmock) CreateModuleDependency(dependentID int64, dependeeID int64) error {
	if d.closed {
		return errors.New("")
	}
	return nil
}

func (d *dbmock) DeleteModuleDependency(dependentID, dependeeID int64) (int64, error) {
	if d.closed {
		return 0, errors.New("")
	}

	return 1, nil
}

func (d *dbmock) DeleteModuleDependencyByDependentID(id int64) (int64, error) {
	if d.closed {
		return 0, errors.New("")
	}

	return 1, nil
}

func (d *dbmock) DeleteModuleDependencyByDependeeID(id int64) (int64, error) {
	if d.closed {
		return 0, errors.New("")
	}

	return 1, nil
}

func (d *dbmock) Close() error {
	d.closed = true
	return nil
}

var (
	db = &dbmock{
		items: []*storage.Item{
			{ID: 1, Value: "httptest", Type: "test", Version: "0.0.1"},
			{ID: 2, Value: "httptest2", Type: "test", Version: "0.0.2"},
		},
		modules: []*storage.Module{
			{ID: 1, Value: "A", Version: "0.0.1"},
			{ID: 2, Value: "B", Version: "0.0.2"},
		},
		itemModules: []*storage.ItemModule{
			{ID: 1, ItemID: 1, ModuleID: 1},
			{ID: 2, ItemID: 2, ModuleID: 2},
		},
		dependencies: []*storage.ModuleDependency{
			{Dependent: 1, Dependee: 2},
		},
	}
)

func reopen() {
	db.closed = false
}

func TestItem(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		input  string
		status int
		err    bool
		closed bool
	}{
		"get item 1":     {input: "1"},
		"missing input":  {input: " ", status: http.StatusNotFound, err: true},
		"wrong input":    {input: "something bad", status: http.StatusNotFound, err: true},
		"closed storage": {input: "1", status: http.StatusInternalServerError, err: true, closed: true},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.closed {
				db.Close()
				defer reopen()
			}

			resp, err := http.Get(fmt.Sprintf("%v/items/%v", srv.URL, tc.input))
			if err != nil {
				t.Fatalf("could not send GET request: %v", err)
			}
			defer resp.Body.Close()

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v %v", tc.status, resp.StatusCode, resp.Status)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got: %v", resp.StatusCode)
			}

			item := &storage.Item{}
			if err := json.NewDecoder(resp.Body).Decode(item); err != nil {
				t.Fatalf("expected a storage.Item, got: %v", err)
			}

			i, _ := strconv.ParseInt(tc.input, 10, 64)

			if item.ID != i {
				t.Fatalf("expected: %v, got: %v", 1, item.ID)
			}
		})
	}
}

func TestItems(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		err bool
	}{
		"get items":      {err: false},
		"closed storage": {err: true},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.err {
				db.Close()
				defer reopen()
			}

			resp, err := http.Get(fmt.Sprintf("%v/items", srv.URL))
			if err != nil {
				t.Fatalf("could not send GET request: %v", err)
			}
			defer resp.Body.Close()

			if tc.err {
				if resp.StatusCode != http.StatusInternalServerError {
					t.Fatalf("expected status Bad Request, got: %v", resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got: %v", resp.StatusCode)
			}

			items := make([]*storage.Item, 0)
			err = json.NewDecoder(resp.Body).Decode(&items)
			if err != nil {
				t.Fatalf("expected slice of storage.Item, got: %v", err)
			}
		})
	}
}

func TestCreateItem(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		input  map[string]interface{}
		status int
		err    bool
		closed bool
	}{
		"correct input": {
			input: map[string]interface{}{
				"value": "httptest", "type": "test", "version": "0.0.1",
			},
		},
		"missing values": {
			input:  nil,
			status: http.StatusBadRequest,
			err:    true,
		},
		"wrong input": {
			input: map[string]interface{}{
				"value": 1,
			},
			status: http.StatusBadRequest,
			err:    true,
		},
		"closed storage": {
			input: map[string]interface{}{
				"value": "httptest", "type": "test", "version": "0.0.1",
			},
			status: http.StatusInternalServerError,
			err:    true,
			closed: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := json.NewEncoder(buf).Encode(tc.input)
			if err != nil {
				t.Fatalf("could not encode input: %v", err)
			}

			if tc.closed {
				db.Close()
				defer reopen()
			}

			resp, err := http.Post(fmt.Sprintf("%v/items", srv.URL), "application/json", buf)
			if err != nil {
				t.Fatalf("could not send Post Request: %v", err)
			}

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v", tc.status, resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("expected status Created, got: %v", resp.StatusCode)
			}

			item := &storage.Item{}
			err = json.NewDecoder(resp.Body).Decode(item)
			if err != nil {
				t.Fatalf("expected a storage.Item, got: %v", err)
			}
		})
	}
}

func TestDeleteItem(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		input  string
		status int
		err    bool
		closed bool
	}{
		"Delete item 1":  {input: "1"},
		"wrong input":    {input: "woop woop", status: http.StatusNotFound, err: true},
		"closed storage": {input: "1", status: http.StatusInternalServerError, err: true, closed: true},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.closed {
				db.Close()
				defer reopen()
			}

			req, _ := http.NewRequest(
				http.MethodDelete, fmt.Sprintf("%v/items/%v", srv.URL, tc.input), nil,
			)

			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Fatalf("could not send Delete Request: %v", err)
			}

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v", tc.status, resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got: %v", resp.StatusCode)
			}

			rows := make(map[string]interface{})
			err = json.NewDecoder(resp.Body).Decode(&rows)
			if err != nil {
				t.Fatalf("expected a map[string]interface{}, got: %v", err)
			}

			deleted := rows["RowsAffected"]

			if deleted != 1.0 {
				t.Fatalf("expected: 1, got: %v", rows["RowsAffected"])
			}
		})
	}
}

func TestModule(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		input  string
		status int
		err    bool
		closed bool
	}{
		"get item 1":     {input: "1"},
		"missing input":  {input: " ", status: http.StatusNotFound, err: true},
		"wrong input":    {input: "something bad", status: http.StatusNotFound, err: true},
		"closed storage": {input: "1", status: http.StatusInternalServerError, err: true, closed: true},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.closed {
				db.Close()
				defer reopen()
			}

			resp, err := http.Get(fmt.Sprintf("%v/modules/%v", srv.URL, tc.input))
			if err != nil {
				t.Fatalf("could not send GET request: %v", err)
			}
			defer resp.Body.Close()

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v %v", tc.status, resp.StatusCode, resp.Status)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got: %v", resp.StatusCode)
			}

			item := &storage.Item{}
			if err := json.NewDecoder(resp.Body).Decode(item); err != nil {
				t.Fatalf("expected a storage.Item, got: %v", err)
			}

			i, _ := strconv.ParseInt(tc.input, 10, 64)

			if item.ID != i {
				t.Fatalf("expected: %v, got: %v", 1, item.ID)
			}
		})
	}
}

func TestModules(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		err bool
	}{
		"get items":      {err: false},
		"closed storage": {err: true},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.err {
				db.Close()
				defer reopen()
			}

			resp, err := http.Get(fmt.Sprintf("%v/modules", srv.URL))
			if err != nil {
				t.Fatalf("could not send GET request: %v", err)
			}
			defer resp.Body.Close()

			if tc.err {
				if resp.StatusCode != http.StatusInternalServerError {
					t.Fatalf("expected status Bad Request, got: %v", resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got: %v", resp.StatusCode)
			}

			items := make([]*storage.Item, 0)
			err = json.NewDecoder(resp.Body).Decode(&items)
			if err != nil {
				t.Fatalf("expected slice of storage.Item, got: %v", err)
			}
		})
	}
}

func TestCreateModule(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		input  map[string]interface{}
		status int
		err    bool
		closed bool
	}{
		"correct input": {
			input: map[string]interface{}{
				"value": "httptest", "version": "0.0.1",
			},
		},
		"missing values": {
			input:  nil,
			status: http.StatusBadRequest,
			err:    true,
		},
		"wrong input": {
			input: map[string]interface{}{
				"value": 1,
			},
			status: http.StatusBadRequest,
			err:    true,
		},
		"closed storage": {
			input: map[string]interface{}{
				"value": "httptest", "version": "0.0.1",
			},
			status: http.StatusInternalServerError,
			err:    true,
			closed: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := json.NewEncoder(buf).Encode(tc.input)
			if err != nil {
				t.Fatalf("could not encode input: %v", err)
			}

			if tc.closed {
				db.Close()
				defer reopen()
			}

			resp, err := http.Post(fmt.Sprintf("%v/modules", srv.URL), "application/json", buf)
			if err != nil {
				t.Fatalf("could not send Post Request: %v", err)
			}

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v", tc.status, resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("expected status Created, got: %v", resp.StatusCode)
			}

			item := &storage.Item{}
			err = json.NewDecoder(resp.Body).Decode(item)
			if err != nil {
				t.Fatalf("expected a storage.Item, got: %v", err)
			}
		})
	}
}

func TestDeleteModule(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		input  string
		status int
		err    bool
		closed bool
	}{
		"Delete module 1": {input: "1"},
		"wrong input": {
			input: "woop woop", status: http.StatusNotFound, err: true,
		},
		"closed storage": {
			input:  "1",
			status: http.StatusInternalServerError,
			err:    true,
			closed: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.closed {
				db.Close()
				defer reopen()
			}

			req, _ := http.NewRequest(
				http.MethodDelete, fmt.Sprintf("%v/modules/%v", srv.URL, tc.input), nil,
			)

			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Fatalf("could not send Delete Request: %v", err)
			}

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v", tc.status, resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got: %v", resp.StatusCode)
			}

			rows := make(map[string]interface{})
			err = json.NewDecoder(resp.Body).Decode(&rows)
			if err != nil {
				t.Fatalf("expected a map[string]interface{}, got: %v", err)
			}

			deleted := rows["RowsAffected"]

			if deleted != 1.0 {
				t.Fatalf("expected: 1, got: %v", rows["RowsAffected"])
			}
		})
	}
}

func TestItemModule(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		input  string
		status int
		err    bool
		closed bool
	}{
		"get item 1":     {input: "1"},
		"missing input":  {input: " ", status: http.StatusNotFound, err: true},
		"wrong input":    {input: "something bad", status: http.StatusNotFound, err: true},
		"closed storage": {input: "1", status: http.StatusInternalServerError, err: true, closed: true},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.closed {
				db.Close()
				defer reopen()
			}

			resp, err := http.Get(fmt.Sprintf("%v/itemmodules/%v", srv.URL, tc.input))
			if err != nil {
				t.Fatalf("could not send GET request: %v", err)
			}
			defer resp.Body.Close()

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v %v", tc.status, resp.StatusCode, resp.Status)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got: %v", resp.StatusCode)
			}

			item := &storage.Item{}
			if err := json.NewDecoder(resp.Body).Decode(item); err != nil {
				t.Fatalf("expected a storage.Item, got: %v", err)
			}

			i, _ := strconv.ParseInt(tc.input, 10, 64)

			if item.ID != i {
				t.Fatalf("expected: %v, got: %v", 1, item.ID)
			}
		})
	}
}

func TestItemModules(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		err bool
	}{
		"get items":      {err: false},
		"closed storage": {err: true},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.err {
				db.Close()
				defer reopen()
			}

			resp, err := http.Get(fmt.Sprintf("%v/itemmodules", srv.URL))
			if err != nil {
				t.Fatalf("could not send GET request: %v", err)
			}
			defer resp.Body.Close()

			if tc.err {
				if resp.StatusCode != http.StatusInternalServerError {
					t.Fatalf("expected status Bad Request, got: %v", resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got: %v", resp.StatusCode)
			}

			items := make([]*storage.Item, 0)
			err = json.NewDecoder(resp.Body).Decode(&items)
			if err != nil {
				t.Fatalf("expected slice of storage.Item, got: %v", err)
			}
		})
	}
}

func TestCreateItemModule(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		input  map[string]interface{}
		status int
		err    bool
		closed bool
	}{
		"correct input": {
			input: map[string]interface{}{
				"item_id": 1, "module_id": 1,
			},
		},
		"missing values": {
			input:  nil,
			status: http.StatusBadRequest,
			err:    true,
		},
		"wrong input": {
			input: map[string]interface{}{
				"value": 1,
			},
			status: http.StatusBadRequest,
			err:    true,
		},
		"closed storage": {
			input: map[string]interface{}{
				"item_id": 1, "module_id": 1,
			},
			status: http.StatusInternalServerError,
			err:    true,
			closed: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := json.NewEncoder(buf).Encode(tc.input)
			if err != nil {
				t.Fatalf("could not encode input: %v", err)
			}

			if tc.closed {
				db.Close()
				defer reopen()
			}

			resp, err := http.Post(fmt.Sprintf("%v/itemmodules", srv.URL), "application/json", buf)
			if err != nil {
				t.Fatalf("could not send Post Request: %v", err)
			}

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v", tc.status, resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("expected status Created, got: %v", resp.StatusCode)
			}

			item := &storage.Item{}
			err = json.NewDecoder(resp.Body).Decode(item)
			if err != nil {
				t.Fatalf("expected a storage.Item, got: %v", err)
			}
		})
	}
}

func TestDeleteItemModule(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		input  string
		status int
		err    bool
		closed bool
	}{
		"Delete module 1": {input: "1"},
		"wrong input": {
			input: "woop woop", status: http.StatusNotFound, err: true,
		},
		"closed storage": {
			input:  "1",
			status: http.StatusInternalServerError,
			err:    true,
			closed: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.closed {
				db.Close()
				defer reopen()
			}

			req, _ := http.NewRequest(
				http.MethodDelete, fmt.Sprintf("%v/itemmodules/%v", srv.URL, tc.input), nil,
			)

			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Fatalf("could not send Delete Request: %v", err)
			}

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v", tc.status, resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got: %v", resp.StatusCode)
			}

			rows := make(map[string]interface{})
			err = json.NewDecoder(resp.Body).Decode(&rows)
			if err != nil {
				t.Fatalf("expected a map[string]interface{}, got: %v", err)
			}

			deleted := rows["RowsAffected"]

			if deleted != 1.0 {
				t.Fatalf("expected: 1, got: %v", rows["RowsAffected"])
			}
		})
	}
}

func TestModuleDependencies(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	moddepURL := "moduledependencies"
	dependentURL := moddepURL + "/dependent"
	dependeeURL := moddepURL + "/dependee"

	tt := map[string]struct {
		input  interface{}
		url    string
		status int
		err    bool
		closed bool
	}{
		"get module dependencies": {
			input: -1,
			url:   moddepURL,
		},
		"get module dependencies by dependent id": {
			input: 1,
			url:   dependentURL,
		},
		"get module dependencies by dependee id": {
			input: 2,
			url:   dependeeURL,
		},
		"wrong dependent id": {
			input:  "test",
			url:    dependentURL,
			status: http.StatusNotFound,
			err:    true,
		},
		"wrong dependee id": {
			input:  "test",
			url:    dependeeURL,
			status: http.StatusNotFound,
			err:    true,
		},
		"closed storage module dependecies": {
			input:  -1,
			url:    moddepURL,
			status: http.StatusInternalServerError,
			err:    true,
			closed: true,
		},
		"closed storage module dependecies by dependent": {
			input:  0,
			url:    dependentURL,
			status: http.StatusInternalServerError,
			err:    true,
			closed: true,
		},
		"closed storage module dependencies by dependee": {
			input:  0,
			url:    dependeeURL,
			status: http.StatusInternalServerError,
			err:    true,
			closed: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.closed {
				db.Close()
				defer reopen()
			}

			url := fmt.Sprintf("%v/%v/%v", srv.URL, tc.url, tc.input)
			if tc.input == -1 {
				url = fmt.Sprintf("%v/%v", srv.URL, tc.url)
			}

			t.Log(url)

			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("could not send GET request: %v", err)
			}
			defer resp.Body.Close()

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v", tc.status, resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got: %v", resp.StatusCode)
			}

			moddeps := make([]*storage.ModuleDependency, 0)
			err = json.NewDecoder(resp.Body).Decode(&moddeps)
			if err != nil {
				t.Fatalf("expected slice of storage.Item, got: %v", err)
			}
		})
	}
}

func TestCreateModuleDependency(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		input  map[string]interface{}
		status int
		err    bool
		closed bool
	}{
		"correct input": {
			input: map[string]interface{}{
				"dependent": 1, "dependee": 1,
			},
		},
		"missing values": {
			input:  nil,
			status: http.StatusBadRequest,
			err:    true,
		},
		"wrong input": {
			input: map[string]interface{}{
				"value": 1,
			},
			status: http.StatusBadRequest,
			err:    true,
		},
		"closed storage": {
			input: map[string]interface{}{
				"dependent": 1, "dependee": 1,
			},
			status: http.StatusInternalServerError,
			err:    true,
			closed: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := json.NewEncoder(buf).Encode(tc.input)
			if err != nil {
				t.Fatalf("could not encode input: %v", err)
			}

			if tc.closed {
				db.Close()
				defer reopen()
			}

			resp, err := http.Post(fmt.Sprintf("%v/moduledependencies", srv.URL), "application/json", buf)
			if err != nil {
				t.Fatalf("could not send Post Request: %v", err)
			}

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v", tc.status, resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("expected status Created, got: %v", resp.StatusCode)
			}

			item := &storage.ModuleDependency{}
			err = json.NewDecoder(resp.Body).Decode(item)
			if err != nil {
				t.Fatalf("expected a storage.Item, got: %v", err)
			}
		})
	}
}

func TestDeleteModuleDependency(t *testing.T) {
	router := New(db)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tt := map[string]struct {
		input  map[string]interface{}
		status int
		url    string
		err    bool
		closed bool
	}{
		"Delete module dependency": {
			input: map[string]interface{}{"dependent": 1, "dependee": 2},
			url:   "moduledependencies/dependent/%v/dependee/%v",
		},
		"Delete module dependency with dependent id": {
			input: map[string]interface{}{"id": 1},
			url:   "moduledependencies/dependent/%v",
		},
		"Delete module dependency with dependee id": {
			input: map[string]interface{}{"id": 2},
			url:   "moduledependencies/dependee/%v",
		},
		"wrong input": {
			input:  map[string]interface{}{"dependent": "w", "dependee": "w"},
			status: http.StatusNotFound,
			url:    "moduledependencies/dependent/%v/dependee/%v",
			err:    true,
		},
		"closed storage": {
			input:  map[string]interface{}{"dependent": 1, "dependee": 2},
			status: http.StatusInternalServerError,
			url:    "moduledependencies/dependent/%v/dependee/%v",
			err:    true,
			closed: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.closed {
				db.Close()
				defer reopen()
			}

			var url string
			if len(tc.input) == 2 {
				url = fmt.Sprintf(tc.url, tc.input["dependent"], tc.input["dependee"])
			} else {
				url = fmt.Sprintf(tc.url, tc.input["id"])
			}

			t.Log(name)
			t.Log(url)

			req, _ := http.NewRequest(
				http.MethodDelete, fmt.Sprintf("%v/%v", srv.URL, url), nil,
			)

			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Fatalf("could not send Delete Request: %v", err)
			}

			if tc.err {
				if resp.StatusCode != tc.status {
					t.Fatalf("expected: %v, got: %v", tc.status, resp.StatusCode)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status OK, got: %v", resp.StatusCode)
			}

			rows := make(map[string]interface{})
			err = json.NewDecoder(resp.Body).Decode(&rows)
			if err != nil {
				t.Fatalf("expected a map[string]interface{}, got: %v", err)
			}

			deleted := rows["RowsAffected"]

			if deleted != 1.0 {
				t.Fatalf("expected: 1, got: %v", rows["RowsAffected"])
			}
		})
	}
}
