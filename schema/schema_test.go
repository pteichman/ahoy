package schema

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/context"
)

func TestSchema_inOrder(t *testing.T) {
	var cur int

	schema := Schema()
	for i, c := range schema {
		if c.Version <= cur {
			t.Fatalf("schemas run backward: schema[%d].Version=%d, was %d", i, c.Version, cur)
		}
	}
}

func TestSchema_Upgrade(t *testing.T) {
	db := newDB(t, 0)
	schema := Schema()

	for _, c := range schema {
		err := Apply(context.Background(), db, schema, c.Version)
		if err != nil {
			t.Fatalf("applying schema %s", c.Name)
		}
	}
}

func newDB(t *testing.T, version int) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("creating test database: %s", err)
	}

	schema := Schema()
	err = Apply(context.Background(), db, schema, version)
	if err != nil {
		t.Fatalf("upgrading schema: %s", err)
	}

	return db
}
