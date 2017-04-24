package postgres

import (
	"database/sql"
	"os"
	"testing"

	"github.com/Boostport/migration"
	"github.com/lib/pq"
)

func TestPostgresDriver(t *testing.T) {

	postgresHost := os.Getenv("POSTGRES_HOST")

	database := "migrationtest"

	for _, useTransactions := range []bool{true, false} {

		// prepare clean database
		connection, err := sql.Open("postgres", "postgres://postgres:@"+postgresHost+"/?sslmode=disable")

		if err != nil {
			t.Fatal(err)
		}

		_, err = connection.Exec("CREATE DATABASE " + database)

		if err != nil {
			t.Fatal(err)
		}

		connection2, err := sql.Open("postgres", "postgres://postgres:@"+postgresHost+"/"+database+"?sslmode=disable")

		if err != nil {
			t.Fatal(err)
		}

		driver, err := New("postgres://postgres:@"+postgresHost+"/"+database+"?sslmode=disable", useTransactions)

		if err != nil {
			t.Errorf("Unable to open connection to postgres server: %s", err)
		}

		migrations := []*migration.PlannedMigration{
			{
				Migration: &migration.Migration{
					ID: "201610041422_init",
					Up: `CREATE TABLE test_table1 (id integer not null primary key);

				     CREATE TABLE test_table2 (id integer not null primary key)
`,
				},
				Direction: migration.Up,
			},
			{
				Migration: &migration.Migration{
					ID:   "201610041425_drop_unused_table",
					Up:   "DROP TABLE test_table2",
					Down: "CREATE TABLE test_table2(id integer not null primary key)",
				},
				Direction: migration.Up,
			},
			{
				Migration: &migration.Migration{
					ID: "201610041422_invalid_sql",
					Up: "CREATE TABLE test_table3 (some error",
				},
				Direction: migration.Up,
			},
		}

		err = driver.Migrate(migrations[0])

		if err != nil {
			t.Errorf("Unexpected error while running migration: %s", err)
		}

		_, err = connection2.Exec("INSERT INTO test_table1 (id) values (1)")

		if err != nil {
			t.Errorf("Unexpected error while testing if migration succeeded: %s", err)
		}

		_, err = connection2.Exec("INSERT INTO test_table2 (id) values (1)")

		if err != nil {
			t.Errorf("Unexpected error while testing if migration succeeded: %s", err)
		}

		err = driver.Migrate(migrations[1])

		if err != nil {
			t.Errorf("Unexpected error while running migration: %s", err)
		}

		if _, err = connection2.Exec("INSERT INTO test_table2 (id) values (1)"); err != nil {
			if err.(*pq.Error).Code.Name() != "undefined_table" {
				t.Errorf("Received an error while inserting into a non-existent table, but it was not a undefined_table error: %s", err)
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

		driver.Close()

		connection2.Close()

		_, err = connection.Exec("DROP DATABASE IF EXISTS " + database)

		if err != nil {
			t.Errorf("Error dropping test database: %s", err)
		}

		connection.Close()
	}
}
