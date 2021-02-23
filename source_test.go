package migration

import (
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/markbates/pkger"
)

func TestGobinDataMigrationSource(t *testing.T) {
	assetMigration := &GoBindataMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "test-migrations",
	}

	driver := getMockDriver()
	applied, err := Migrate(driver, assetMigration, Up, 0)
	if err != nil {
		t.Errorf("Unexpected error while performing go-bindata migration: %s", err)
	}
	if applied != 3 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 3, applied)
	}
	if len(driver.applied) != 3 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied, len(driver.applied))
	}
}

func TestPackrMigrationSource(t *testing.T) {
	assetMigration := &PackrMigrationSource{
		Box: packr.New("migrations", "."),
		Dir: "test-migrations",
	}

	driver := getMockDriver()
	applied, err := Migrate(driver, assetMigration, Up, 0)
	if err != nil {
		t.Errorf("Unexpected error while performing packr migration: %s", err)
	}
	if applied != 3 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 3, applied)
	}
	if len(driver.applied) != 3 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied, len(driver.applied))
	}
}

func TestPackrMigrationSourceWithoutDir(t *testing.T) {
	assetMigration := &PackrMigrationSource{
		Box: packr.New("test-migrations", "test-migrations"),
	}

	driver := getMockDriver()
	applied, err := Migrate(driver, assetMigration, Up, 0)
	if err != nil {
		t.Errorf("Unexpected error while performing packr migration: %s", err)
	}
	if applied != 3 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 3, applied)
	}
	if len(driver.applied) != 3 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied, len(driver.applied))
	}
}

func TestPkgerMigrationSource(t *testing.T) {

	dir := pkger.Include("/test-migrations")

	assetMigration := &PkgerMigrationSource{
		Dir: dir,
	}

	driver := getMockDriver()
	applied, err := Migrate(driver, assetMigration, Up, 0)
	if err != nil {
		t.Errorf("Unexpected error while performing packr migration: %s", err)
	}
	if applied != 3 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 3, applied)
	}
	if len(driver.applied) != 3 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied, len(driver.applied))
	}
}
