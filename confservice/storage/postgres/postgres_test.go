package postgres

import (
	"testing"

	"github.com/Glorforidor/conmansys/confservice/storage"
)

var p *postgres

// there must be a better way?
func setup(t *testing.T) {
	host := "172.17.0.2"
	port := "5432"
	user := "postgres"
	pass := "secret"
	dbname := "conmansys"

	pp, err := New(host, port, user, pass, dbname)
	if err != nil {
		t.Fatal(err)
	}

	p = pp.(*postgres)
}

func cleanup(t *testing.T) {
	err := p.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetItem(t *testing.T) {
	setup(t)
	defer cleanup(t)

	tests := map[string]struct {
		id   int64
		want *storage.Item
		err  bool
	}{
		"fetch item 1":          {id: 1, want: &storage.Item{1, "tax_flow", "window", "1.0.0"}},
		"fetch item 2":          {id: 2, want: &storage.Item{2, "tax", "domain", "1.0.0"}},
		"item 3 does not exist": {id: 3, want: nil, err: true},
	}

	for name, test := range tests {
		got, err := p.GetItem(test.id)
		if err != nil {
			if test.err {
				continue
			}
			t.Fatal(err)
		}
		if *got != *test.want {
			t.Fatalf("%v: expected: %v, got: %v", name, test.want, got)
		}
	}
}

func TestGetItems(t *testing.T) {
	setup(t)
	defer cleanup(t)

	_, err := p.GetItems()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateItem(t *testing.T) {
	setup(t)
	defer cleanup(t)
}
