package database

import (
	"path/filepath"
	"testing"
)

func TestNewConnectionRejectsUnknownType(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "x.db")
	for _, typ := range []string{"", "sqlite3", "SQLite", "postgresql"} {
		if _, err := NewConnection(typ, dsn); err == nil {
			t.Errorf("dbtype %q: expected error", typ)
		}
	}
	if _, err := NewConnection("sqlite", dsn); err != nil {
		t.Fatalf("sqlite: %v", err)
	}
}
