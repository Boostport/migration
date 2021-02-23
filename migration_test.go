package migration

import (
	"reflect"
	"sort"
	"testing"
)

func TestDirectionString(t *testing.T) {
	if Up.String() != "up" {
		t.Errorf("Expect `Up` to be 'up', got '%s'", Up.String())
	}
	if Down.String() != "down" {
		t.Errorf("Expect `Down` to be 'down', got '%s'", Down.String())
	}
	var Other Direction = -1
	if Other.String() != "directionless" {
		t.Errorf("Expect `Other` to be 'directionless', got '%s'", Other.String())
	}
}

func TestMigrationSorting(t *testing.T) {
	unsorted := []*Migration{
		{
			ID: "1475461906_remove_name_column",
		},
		{
			ID: "1375461906_init",
		},
		{
			ID: "1575461906_remove_users_table",
		},
		{
			ID: "1475461916_add_sales_table",
		},
		{
			ID: "1475461904442_remove_subscriptions_table",
		},
		{
			ID: "1475461904_add_last_name_column",
		},
	}

	sorted := []*Migration{
		{
			ID: "1375461906_init",
		},
		{
			ID: "1475461904_add_last_name_column",
		},
		{
			ID: "1475461906_remove_name_column",
		},
		{
			ID: "1475461916_add_sales_table",
		},
		{
			ID: "1575461906_remove_users_table",
		},
		{
			ID: "1475461904442_remove_subscriptions_table",
		},
	}

	sort.Sort(byID(unsorted))

	if !reflect.DeepEqual(unsorted, sorted) {
		t.Error("Sorted migrations are not in the correct order.")
	}
}

func TestMigrationSortingWithNonNumericIds(t *testing.T) {
	unsorted := []*Migration{
		{
			ID: "b_init",
		},
		{
			ID: "a_remove_users_table",
		},
		{
			ID: "d_remove_users_table",
		},
		{
			ID: "147546_add_sales_table",
		},
		{
			ID: "c_remove_users_table",
		},
		{
			ID: "1_remove_name_column",
		},
	}

	sorted := []*Migration{
		{
			ID: "1_remove_name_column",
		},
		{
			ID: "147546_add_sales_table",
		},
		{
			ID: "a_remove_users_table",
		},
		{
			ID: "b_init",
		},
		{
			ID: "c_remove_users_table",
		},
		{
			ID: "d_remove_users_table",
		},
	}

	sort.Sort(byID(unsorted))

	if !reflect.DeepEqual(unsorted, sorted) {
		t.Error("Sorted migrations are not in the correct order.")
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
		t.Error("Expected error while running migration, but there was no error")
	}
	if applied != 1 {
		t.Errorf("%d migrations should be applied, but %d was applied.", 1, applied)
	}

	applied2, err := Migrate(driver, memoryMigration, Down, 1)
	if err == nil {
		t.Error("Expected error while running migration, but there was no error")
	}
	if applied2 != 0 {
		t.Errorf("No migrations should be applied, but %d was applied.", applied2)
	}
}
