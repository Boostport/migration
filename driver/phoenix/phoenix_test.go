package phoenix

import (
	"database/sql"
	"os"
	"testing"

	"github.com/Boostport/migration"
	"github.com/Boostport/migration/parser"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/apache/calcite-avatica-go/v5/errors"
)

func TestPhoenixDriver(t *testing.T) {
	phoenixHost := os.Getenv("PHOENIX_HOST")

	// prepare clean database
	connection, err := sql.Open("avatica", phoenixHost)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := connection.Close()
		if err != nil {
			t.Errorf("unexpected error while closing the phoenix connection: %v", err)
		}
	}()

	schema := "migrationtest"
	_, err = connection.Exec("CREATE SCHEMA IF NOT EXISTS " + schema)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := New(phoenixHost + "/")
	if err != nil {
		t.Errorf("Unable to open connection to phoenix server: %s", err)
	}

	defer func() {
		err := driver.Close()
		if err != nil {
			t.Errorf("unexpected error while closing the phoenix driver: %v", err)
		}
	}()

	defer func() {
		_, err := connection.Exec("DROP TABLE IF EXISTS test_table1")
		if err != nil {
			t.Errorf("unexpected error while dropping the phoenix table: %v", err)
		}
		_, err = connection.Exec("DROP TABLE IF EXISTS test_table2")
		if err != nil {
			t.Errorf("unexpected error while dropping the phoenix table: %v", err)
		}
		_, err = connection.Exec("DROP TABLE IF EXISTS schema_migration")
		if err != nil {
			t.Errorf("unexpected error while dropping the phoenix table: %v", err)
		}
		_, err = connection.Exec("DROP SCHEMA IF EXISTS " + schema)
		if err != nil {
			t.Errorf("unexpected error while dropping the phoenix schema %s: %v", schema, err)
		}
	}()

	migrations := []*migration.PlannedMigration{
		{
			Migration: &migration.Migration{
				ID: "201610041422_init",
				Up: &parser.ParsedMigration{
					Statements: []string{
						`CREATE TABLE test_table1 (id integer not null primary key);

				   		 CREATE TABLE test_table2 (id integer not null primary key)`,
					},
					UseTransaction: false,
				},
			},
			Direction: migration.Up,
		},
		{
			Migration: &migration.Migration{
				ID: "201610041425_drop_unused_table",
				Up: &parser.ParsedMigration{
					Statements: []string{
						"DROP TABLE test_table2",
					},
					UseTransaction: false,
				},
				Down: &parser.ParsedMigration{
					Statements: []string{
						"CREATE TABLE test_table2(id integer not null primary key)",
					},
					UseTransaction: false,
				},
			},
			Direction: migration.Up,
		},
		{
			Migration: &migration.Migration{
				ID: "201610041422_invalid_sql",
				Up: &parser.ParsedMigration{
					Statements: []string{
						"CREATE TABLE test_table3 (some error",
					},
					UseTransaction: false,
				},
			},
			Direction: migration.Up,
		},
	}

	err = driver.Migrate(migrations[0])
	if err != nil {
		t.Errorf("Unexpected error while running migration: %s", err)
	}

	_, err = connection.Exec("UPSERT INTO test_table1 (id) values (1)")
	if err != nil {
		t.Errorf("Unexpected error while testing if migration succeeded: %s", err)
	}

	_, err = connection.Exec("UPSERT INTO test_table2 (id) values (1)")
	if err != nil {
		t.Errorf("Unexpected error while testing if migration succeeded: %s", err)
	}

	err = driver.Migrate(migrations[1])
	if err != nil {
		t.Errorf("Unexpected error while running migration: %s", err)
	}

	if _, err = connection.Exec("UPSERT INTO test_table2 (id) values (1)"); err != nil {
		if err.(errors.ResponseError).Name != "table_undefined" {
			t.Errorf("Received an error while inserting into a non-existent table, but it was not a table_undefined error: %s", err)
		}
	} else {
		t.Error("Expected an error while inserting into non-existent table, but did not receive any.")
	}

	err = driver.Migrate(migrations[2])
	if err == nil {
		t.Error("Expected an error while executing invalid statement, but did not receive any.")
	}

	versions, err := driver.Versions()
	if err != nil {
		t.Errorf("Unexpected error while retriving version information: %s", err)
	}
	if len(versions) != 2 {
		t.Errorf("Expected %d versions to be applied, %d was actually applied.", 2, len(versions))
	}

	migrations[1].Direction = migration.Down
	err = driver.Migrate(migrations[1])
	if err != nil {
		t.Errorf("Unexpected error while running migration: %s", err)
	}

	versions, err = driver.Versions()
	if err != nil {
		t.Errorf("Unexpected error while retriving version information: %s", err)
	}
	if len(versions) != 1 {
		t.Errorf("Expected %d versions to be applied, %d was actually applied.", 2, len(versions))
	}
}

func TestCreateDriverUsingInvalidDBInstance(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error opening stub database connection: %s", err)
	}

	_, err = NewFromDB(db)
	if err == nil {
		t.Error("Expected error when creating Phoenix driver with a non-Phoenix database instance, but there was no error")
	}
}

func TestCreateDriverUsingDBInstance(t *testing.T) {
	phoenixHost := os.Getenv("PHOENIX_HOST")

	// prepare clean database
	connection, err := sql.Open("avatica", phoenixHost)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := connection.Close()
		if err != nil {
			t.Errorf("unexpected error while closing the phoenix connection: %v", err)
		}
	}()

	schema := "migrationtest"
	_, err = connection.Exec("CREATE SCHEMA IF NOT EXISTS " + schema)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_, err := connection.Exec("DROP TABLE IF EXISTS schema_migration")
		if err != nil {
			t.Errorf("unexpected error while dropping the phoenix table: %v", err)
		}
		_, err = connection.Exec("DROP SCHEMA IF EXISTS " + schema)
		if err != nil {
			t.Errorf("unexpected error while dropping the phoenix schema %s: %v", schema, err)
		}
	}()

	db, err := sql.Open("avatica", phoenixHost+"/")
	if err != nil {
		t.Fatalf("Could not open avatica connection: %s", err)
	}

	driver, err := NewFromDB(db)
	if err != nil {
		t.Errorf("Unable to create Avatica driver: %s", err)
	}
	defer func() {
		err := driver.Close()
		if err != nil {
			t.Errorf("unexpected error while closing the phoenix driver: %v", err)
		}
	}()
}
