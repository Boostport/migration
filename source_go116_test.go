//go:build go1.16
// +build go1.16

package migration

import (
	"embed"
	"testing"
)

//go:embed test-migrations
var embedFS embed.FS

func TestEmbedMigrationSource(t *testing.T) {

	assetMigration := &EmbedMigrationSource{
		EmbedFS: embedFS,
		Dir:     "test-migrations",
	}

	driver := getMockDriver()
	applied, err := Migrate(driver, assetMigration, Up, 0)
	if err != nil {
		t.Errorf("Unexpected error while performing embed migration: %s", err)
	}
	if applied != 3 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 3, applied)
	}
	if len(driver.applied) != 3 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied, len(driver.applied))
	}
}
