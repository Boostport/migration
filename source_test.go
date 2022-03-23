package migration

import (
	"testing"
)

func TestGolangMigrationSource(t *testing.T) {

	assetMigration := NewGolangMigrationSource()

	assetMigration.AddMigration("1_init", Up, func() error {
		return nil
	})

	assetMigration.AddMigration("2_update", Up, func() error {
		return nil
	})

	assetMigration.AddMigration("3_add_column", Up, func() error {
		return nil
	})

	driver := getMockDriver()
	applied, err := Migrate(driver, assetMigration, Up, 0)
	if err != nil {
		t.Errorf("Unexpected error while performing golang migration: %s", err)
	}
	if applied != 3 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 3, applied)
	}
	if len(driver.applied) != 3 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied, len(driver.applied))
	}
}
