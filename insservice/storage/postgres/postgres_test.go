package postgres

import (
	"testing"

	"github.com/Glorforidor/conmansys/insservice/storage"
)

const (
	table = `
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
);`

	insert = `
INSERT INTO conf_item (conf_item_value, conf_item_type, conf_item_version) VALUES
('tax_income_window', 'window', '1.0.0'),
('tax', 'domain', '1.0.0'),
('payment_window', 'window', '1.0.0'),
('payment', 'domain', '1.0.0'),
('refund_window', 'window', '2.0.0'),
('refund', 'domain', '0.0.1'),
('management_tax_window', 'window', '2.0.0'),
('mangement', 'domain', '0.0.2');

INSERT INTO conf_module (conf_module_value, conf_module_version) VALUES
('A', '0.0.10'),
('B', '0.0.11'),
('C', '0.0.12'),
('D', '0.0.13'),
('E', '0.0.14'),
('F', '0.0.15');

INSERT INTO conf_item_module (conf_item_id, conf_module_id) VALUES
(1, 1),
(2, 1),
(3, 2),
(4, 2),
(5, 3),
(6, 3),
(7, 4),
(8, 4);

INSERT INTO conf_module_dependency (dependent, dependee) VALUES
(4, 1),
(4, 2),
(4, 3),
(5, 4),
(6, 1),
(6, 5);
`
)

func setup(t *testing.T) *postgres {

	p, err := New(
		"172.19.0.2",
		"5432",
		"postgres",
		"secret",
		"postgres", // connect to default database
	)
	if err != nil {
		t.Fatalf("could not create postgres connection: %v", err)
	}

	_, err = p.db.Exec(table)
	if err != nil {
		t.Fatalf("could not create tables: %v", err)
	}

	_, err = p.db.Exec(insert)
	if err != nil {
		t.Fatalf("could not insert data into tables: %v", err)
	}

	return p
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
			host:   "172.19.0.2", // important a test database is running
			port:   "5432",
			user:   "postgres",
			pass:   "secret",
			dbname: "postgres",
		},
		"bad connection": {
			host:   "172.19.0.2",
			port:   "5432",
			user:   "postgres",
			pass:   "secret",
			dbname: "testtest", // database does not exist
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

func TestGetItems(t *testing.T) {
	p := setup(t)

	tt := map[string]struct {
		modules []storage.Module
		want    map[string]bool
	}{
		"items for module 4 and 6": {
			modules: []storage.Module{
				{ID: 4},
				{ID: 6},
			},
			want: map[string]bool{
				"tax_income_window":     true,
				"tax":                   true,
				"payment_window":        true,
				"payment":               true,
				"refund_window":         true,
				"refund":                true,
				"management_tax_window": true,
				"mangement":             true,
			},
		},
		"items for module 6": {
			modules: []storage.Module{
				{ID: 6},
			},
			want: map[string]bool{
				"tax_income_window":     true,
				"tax":                   true,
				"payment_window":        true,
				"payment":               true,
				"refund_window":         true,
				"refund":                true,
				"management_tax_window": true,
				"mangement":             true,
			},
		},
		"items for module 4": {
			modules: []storage.Module{
				{ID: 4},
			},
			want: map[string]bool{
				"tax_income_window":     true,
				"tax":                   true,
				"payment_window":        true,
				"payment":               true,
				"refund_window":         true,
				"refund":                true,
				"management_tax_window": true,
				"mangement":             true,
			},
		},
		"items for module 5": {
			modules: []storage.Module{
				{ID: 5},
			},
			want: map[string]bool{
				"tax_income_window":     true,
				"tax":                   true,
				"payment_window":        true,
				"payment":               true,
				"refund_window":         true,
				"refund":                true,
				"management_tax_window": true,
				"mangement":             true,
			},
		},
		"items for module 1, 2 and 3": {
			modules: []storage.Module{
				{ID: 1},
				{ID: 2},
				{ID: 3},
			},
			want: map[string]bool{
				"tax_income_window": true,
				"tax":               true,
				"payment_window":    true,
				"payment":           true,
				"refund_window":     true,
				"refund":            true,
			},
		},
		"does not exist": {
			modules: []storage.Module{
				{},
			},
			want: map[string]bool{},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			items, err := p.GetItems(tc.modules...)
			if err != nil {
				t.Fatal(err)
			}

			if len(items) != len(tc.want) {
				t.Log(items)
				t.Errorf("expected map length of: %v, got: %v", len(tc.want), len(items))
			}

			for k := range tc.want {
				_, ok := items[k]
				if !ok {
					t.Errorf("expected key: %v in:  %v", k, items)
				}
			}
		})
	}
}
