package golang

import (
	"errors"
	"testing"

	"github.com/Boostport/migration"
)

func TestGolangDriver(t *testing.T) {

	mockDB := map[string]string{}

	mockDBVersions := map[string]struct{}{}

	config := NewGolangConfig()
	config.Set("test", "test")
	config.Set("db", mockDB)
	config.Set("versions", mockDBVersions)

	if v := config.Get("test"); v != "test" {
		t.Errorf("Expected test key in config to return %s, got %v", "test", v)
	}

	source := NewGolangSource()

	source.AddMigration("1_init", migration.Up, func(c *golangConfig) error {

		db := c.Get("db")

		if db == nil {
			return errors.New("db is not in config map")
		}

		db.(map[string]string)["v1"] = "test"

		return nil
	})

	source.AddMigration("1_init", migration.Down, func(c *golangConfig) error {

		db := c.Get("db")

		if db == nil {
			return errors.New("db is not in config map")
		}

		delete(db.(map[string]string), "v1")

		return nil
	})

	source.AddMigration("2_update", migration.Up, func(c *golangConfig) error {

		db := c.Get("db")

		if db == nil {
			return errors.New("db is not in config map")
		}

		db.(map[string]string)["v2"] = "test"

		return nil
	})

	source.AddMigration("2_update", migration.Down, func(c *golangConfig) error {

		db := c.Get("db")

		if db == nil {
			return errors.New("db is not in config map")
		}

		delete(db.(map[string]string), "v2")

		return nil
	})

	applied := func(c *golangConfig) ([]string, error) {
		versions := c.Get("versions")

		if versions == nil {
			return []string{}, errors.New("versions is not in config map")
		}

		keys := []string{}

		for version := range versions.(map[string]struct{}) {
			keys = append(keys, version)
		}

		return keys, nil
	}

	updateVersion := func(id string, direction migration.Direction, c *golangConfig) error {
		versions := c.Get("versions")

		if versions == nil {
			return errors.New("versions is not in config map")
		}

		if direction == migration.Up {
			versions.(map[string]struct{})[id] = struct{}{}
		} else if direction == migration.Down {
			delete(versions.(map[string]struct{}), id)
		}

		return nil
	}

	driver, err := NewGolang(source, updateVersion, applied, config)

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
