package golang

import (
	"testing"

	"github.com/Boostport/migration"
)

func TestGolangDriver(t *testing.T) {

	mockDB := map[string]string{}

	mockDBVersions := map[string]struct{}{}

	source := NewSource()

	source.AddMigration("1_init", migration.Up, func() error {

		mockDB["v1"] = "test"

		return nil
	})

	source.AddMigration("1_init", migration.Down, func() error {

		delete(mockDB, "v1")

		return nil
	})

	source.AddMigration("2_update", migration.Up, func() error {

		mockDB["v2"] = "test"

		return nil
	})

	source.AddMigration("2_update", migration.Down, func() error {

		delete(mockDB, "v2")

		return nil
	})

	applied := func() ([]string, error) {

		var keys []string

		for version := range mockDBVersions {
			keys = append(keys, version)
		}

		return keys, nil
	}

	updateVersion := func(id string, direction migration.Direction) error {

		if direction == migration.Up {
			mockDBVersions[id] = struct{}{}
		} else if direction == migration.Down {
			delete(mockDBVersions, id)
		}

		return nil
	}

	driver, err := New(source, updateVersion, applied)

	if err != nil {
		t.Errorf("Unexpected error while creating driver: %s", err)
	}

	count, err := migration.Migrate(driver, source, migration.Up, 1)

	if err != nil {
		t.Errorf("Unexpected error while running up migration: %s", err)
	}

	if count != 1 {
		t.Errorf("Expected %d migrations to be run, %d was actually run", 1, count)
	}

	if mockDB["v1"] != "test" {
		t.Error("Migration was not run correctly")
	}

	if _, ok := mockDBVersions["1_init"]; !ok {
		t.Error("Version was not inserted correctly")
	}

	count, err = migration.Migrate(driver, source, migration.Down, 1)

	if err != nil {
		t.Errorf("Unexpected error while running up migration: %s", err)
	}

	if count != 1 {
		t.Errorf("Expected %d migrations to be run, %d was actually run", 1, count)
	}

	if mockDB["v1"] != "" {
		t.Error("Migration was not run correctly")
	}

	if _, ok := mockDBVersions["1_init"]; ok {
		t.Error("Version was not deleted correctly")
	}

	count, err = migration.Migrate(driver, source, migration.Up, 0)

	if err != nil {
		t.Errorf("Unexpected error while running up migration: %s", err)
	}

	if count != 2 {
		t.Errorf("Expected %d migrations to be run, %d was actually run", 2, count)
	}

	if mockDB["v1"] != "test" {
		t.Error("Migration was not run correctly")
	}

	if mockDB["v2"] != "test" {
		t.Error("Migration was not run correctly")
	}

	if _, ok := mockDBVersions["1_init"]; !ok {
		t.Error("Version was not inserted correctly")
	}

	if _, ok := mockDBVersions["2_update"]; !ok {
		t.Error("Version was not inserted correctly")
	}

	count, err = migration.Migrate(driver, source, migration.Down, 0)

	if err != nil {
		t.Errorf("Unexpected error while running up migration: %s", err)
	}

	if count != 2 {
		t.Errorf("Expected %d migrations to be run, %d was actually run", 2, count)
	}

	if mockDB["v1"] != "" {
		t.Error("Migration was not run correctly")
	}

	if mockDB["v2"] != "" {
		t.Error("Migration was not run correctly")
	}

	if _, ok := mockDBVersions["1_init"]; ok {
		t.Error("Version was not deleted correctly")
	}

	if _, ok := mockDBVersions["2_update"]; ok {
		t.Error("Version was not deleted correctly")
	}
}
