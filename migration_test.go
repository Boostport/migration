package migration

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type mockDriver struct {
	applied []string
}

func (m *mockDriver) Close() error {
	return nil
}

func (m *mockDriver) Migrate(migration *PlannedMigration) error {

	var content string

	if migration.Direction == Up {
		content = migration.Up
	} else {
		content = migration.Down
	}

	if strings.Contains(content, "error") {
		return errors.New("Error executing migration.")
	}

	versionIndex := -1

	for i, version := range m.applied {
		if version == migration.Id {
			versionIndex = i
			break
		}
	}

	if migration.Direction == Up {
		if versionIndex == -1 {
			m.applied = append(m.applied, migration.Id)
		}
	} else {
		if versionIndex != -1 {
			m.applied = append(m.applied[:versionIndex], m.applied[versionIndex+1:]...)
		}
	}

	return nil
}

func (m *mockDriver) Versions() ([]string, error) {
	return m.applied, nil
}

func getMockDriver() *mockDriver {
	return &mockDriver{
		applied: []string{},
	}
}

func TestMigrationSorting(t *testing.T) {

	unsorted := []*Migration{
		&Migration{
			Id: "1475461906_remove_name_column",
		},
		&Migration{
			Id: "1375461906_init",
		},
		&Migration{
			Id: "1575461906_remove_users_table",
		},
		&Migration{
			Id: "1475461916_add_sales_table",
		},
		&Migration{
			Id: "1475461904442_remove_subscriptions_table",
		},
		&Migration{
			Id: "1475461904_add_last_name_column",
		},
	}

	sorted := []*Migration{
		&Migration{
			Id: "1375461906_init",
		},
		&Migration{
			Id: "1475461904_add_last_name_column",
		},
		&Migration{
			Id: "1475461906_remove_name_column",
		},
		&Migration{
			Id: "1475461916_add_sales_table",
		},
		&Migration{
			Id: "1575461906_remove_users_table",
		},
		&Migration{
			Id: "1475461904442_remove_subscriptions_table",
		},
	}

	sort.Sort(byId(unsorted))

	if !reflect.DeepEqual(unsorted, sorted) {
		t.Error("Sorted migrations are not in the correct order.")
	}
}

func TestMigrationSortingWithNonNumericIds(t *testing.T) {

	unsorted := []*Migration{
		&Migration{
			Id: "b_init",
		},
		&Migration{
			Id: "a_remove_users_table",
		},
		&Migration{
			Id: "d_remove_users_table",
		},
		&Migration{
			Id: "147546_add_sales_table",
		},
		&Migration{
			Id: "c_remove_users_table",
		},
		&Migration{
			Id: "1_remove_name_column",
		},
	}

	sorted := []*Migration{
		&Migration{
			Id: "1_remove_name_column",
		},
		&Migration{
			Id: "147546_add_sales_table",
		},
		&Migration{
			Id: "a_remove_users_table",
		},
		&Migration{
			Id: "b_init",
		},
		&Migration{
			Id: "c_remove_users_table",
		},
		&Migration{
			Id: "d_remove_users_table",
		},
	}

	sort.Sort(byId(unsorted))

	if !reflect.DeepEqual(unsorted, sorted) {
		t.Errorf("Sorted migrations are not in the correct order.")
	}
}

func TestAssetMigrationSource(t *testing.T) {

	assetMigration := &AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "test-migrations",
	}

	driver := getMockDriver()

	applied, err := Migrate(driver, assetMigration, Up, 0)

	if err != nil {
		t.Errorf("Unexpected error while performing asset migration: %s", err)
	}

	if applied != 3 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 3, applied)
	}

	if len(driver.applied) != 3 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied, len(driver.applied))
	}
}

func TestMigrationWithHoles(t *testing.T) {

	memoryMigration := &MemoryMigrationSource{
		Files: map[string]string{
			"1_init.up.sql":            "",
			"1_init.down.sql":          "",
			"3_second_update.up.sql":   "",
			"3_second_update.down.sql": "",
		},
	}

	driver := getMockDriver()

	applied, err := Migrate(driver, memoryMigration, Up, 0)

	if err != nil {
		t.Errorf("Unexpected error while performing asset migration: %s", err)
	}

	if applied != 2 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 2, applied)
	}

	if len(driver.applied) != 2 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied, len(driver.applied))
	}

	// Add migrations
	memoryMigration.Files["2_first_update.up.sql"] = ""
	memoryMigration.Files["2_first_update.down.sql"] = ""
	memoryMigration.Files["4_another_update.up.sql"] = ""
	memoryMigration.Files["4_another_update.up.sql"] = ""

	applied2, err := Migrate(driver, memoryMigration, Up, 0)

	if err != nil {
		t.Errorf("Unexpected error while performing asset migration: %s", err)
	}

	if applied2 != 2 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 2, applied2)
	}

	if len(driver.applied) != 4 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied2, len(driver.applied))
	}
}

func TestMigrateUpWithLimit(t *testing.T) {

	memoryMigration := &MemoryMigrationSource{
		Files: map[string]string{
			"1_init.up.sql":             "",
			"1_init.down.sql":           "",
			"2_first_update.up.sql":     "",
			"2_first_update.down.sql":   "",
			"3_second_update.up.sql":    "",
			"3_second_update.down.sql":  "",
			"4_another_update.up.sql":   "",
			"4_another_update.down.sql": "",
		},
	}

	driver := getMockDriver()

	applied, err := Migrate(driver, memoryMigration, Up, 2)

	if err != nil {
		t.Errorf("Unexpected error while performing asset migration: %s", err)
	}

	if applied != 2 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 2, applied)
	}

	if len(driver.applied) != 2 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied, len(driver.applied))
	}

	applied2, err := Migrate(driver, memoryMigration, Up, 2)

	if err != nil {
		t.Errorf("Unexpected error while performing asset migration: %s", err)
	}

	if applied2 != 2 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 2, applied2)
	}

	if len(driver.applied) != 4 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied2, len(driver.applied))
	}
}

func TestMigrateDownWithLimit(t *testing.T) {

	memoryMigration := &MemoryMigrationSource{
		Files: map[string]string{
			"1_init.up.sql":             "",
			"1_init.down.sql":           "",
			"2_first_update.up.sql":     "",
			"2_first_update.down.sql":   "",
			"3_second_update.up.sql":    "",
			"3_second_update.down.sql":  "",
			"4_another_update.up.sql":   "",
			"4_another_update.down.sql": "",
		},
	}

	driver := getMockDriver()

	applied, err := Migrate(driver, memoryMigration, Up, 0)

	if err != nil {
		t.Errorf("Unexpected error while performing asset migration: %s", err)
	}

	if applied != 4 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 4, applied)
	}

	if len(driver.applied) != 4 {
		t.Errorf("Applied %d migrations, but driver is showing %d applied.", applied, len(driver.applied))
	}

	applied2, err := Migrate(driver, memoryMigration, Down, 2)

	if err != nil {
		t.Errorf("Unexpected error while performing asset migration: %s", err)
	}

	if applied2 != 2 {
		t.Errorf("Expected %d migrations to be applied, %d applied.", 2, applied2)
	}

	if len(driver.applied) != 2 {
		t.Errorf("There should only be %d migrations after migrating down", 2)
	}
}

func TestMigrationWithError(t *testing.T) {

	memoryMigration := &MemoryMigrationSource{
		Files: map[string]string{
			"1_init.up.sql":     "",
			"1_init.down.sql":   "error",
			"2_update.up.sql":   "error",
			"2_update.down.sql": "",
		},
	}

	driver := getMockDriver()

	applied, err := Migrate(driver, memoryMigration, Up, 2)

	if err == nil {
		t.Errorf("Expected error while running migration, but there was no error")
	}

	if applied != 1 {
		t.Errorf("%d migrations should be applied, but %d was applied.", 1, applied)
	}

	applied2, err := Migrate(driver, memoryMigration, Down, 1)

	if err == nil {
		t.Errorf("Expected error while running migration, but there was no error")
	}

	if applied2 != 0 {
		t.Errorf("No migrations should be applied, but %d was applied.", applied2)
	}
}
