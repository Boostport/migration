package postgres

import (
	"database/sql"
	"os"
	"testing"

	"github.com/Boostport/migration"
	"github.com/Boostport/migration/parser"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func TestPostgresDriver(t *testing.T) {
	postgresHost := os.Getenv("POSTGRES_HOST")

	database := "migrationtest"

	// prepare clean database
	connection, err := sql.Open("postgres", "postgres://postgres:@"+postgresHost+"/?sslmode=disable")

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := connection.Close()
		if err != nil {
			t.Errorf("unexpected error while closing the postgres connection: %v", err)
		}
	}()

	_, err = connection.Exec("CREATE DATABASE " + database)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_, err := connection.Exec("DROP DATABASE IF EXISTS " + database)
		if err != nil {
			t.Errorf("unexpected error while dropping the postgres database %s: %v", database, err)
		}
	}()

	connection2, err := sql.Open("postgres", "postgres://postgres:@"+postgresHost+"/"+database+"?sslmode=disable")

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := connection2.Close()
		if err != nil {
			t.Errorf("unexpected error while closing the postgres connection: %v", err)
		}
	}()

	driver, err := New("postgres://postgres:@" + postgresHost + "/" + database + "?sslmode=disable")

	if err != nil {
		t.Errorf("unable to open connection to postgres server: %s", err)
	}

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
		t.Errorf("unexpected error while running migration: %s", err)
	}

	_, err = connection2.Exec("INSERT INTO test_table1 (id) values (1)")

	if err != nil {
		t.Errorf("unexpected error while testing if migration succeeded: %s", err)
	}

	_, err = connection2.Exec("INSERT INTO test_table2 (id) values (1)")

	if err != nil {
		t.Errorf("unexpected error while testing if migration succeeded: %s", err)
	}

	err = driver.Migrate(migrations[1])

	if err != nil {
		t.Errorf("unexpected error while running migration: %s", err)
	}

	if _, err = connection2.Exec("INSERT INTO test_table2 (id) values (1)"); err != nil {
		if err.(*pq.Error).Code.Name() != "undefined_table" {
			t.Errorf("received an error while inserting into a non-existent table, but it was not a undefined_table error: %s", err)
		}
	} else {
		t.Error("expected an error while inserting into non-existent table, but did not receive any.")
	}

	err = driver.Migrate(migrations[2])

	if err == nil {
		t.Error("expected an error while executing invalid statement, but did not receive any.")
	}

	versions, err := driver.Versions()

	if err != nil {
		t.Errorf("unexpected error while retriving version information: %s", err)
	}

	if len(versions) != 2 {
		t.Errorf("expected %d versions to be applied, %d was actually applied.", 2, len(versions))
	}

	migrations[1].Direction = migration.Down

	err = driver.Migrate(migrations[1])

	if err != nil {
		t.Errorf("unexpected error while running migration: %s", err)
	}

	versions, err = driver.Versions()

	if err != nil {
		t.Errorf("unexpected error while retriving version information: %s", err)
	}

	if len(versions) != 1 {
		t.Errorf("expected %d versions to be applied, %d was actually applied.", 2, len(versions))
	}

	err = driver.Close()
	if err != nil {
		t.Errorf("unexpected error %v while closing the postgres driver.", err)
	}
}

func TestCreateDriverUsingInvalidDBInstance(t *testing.T) {
	db, _, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error opening stub database connection: %s", err)
	}

	_, err = NewFromDB(db)

	if err == nil {
		t.Error("expected error when creating Postgres driver with a non-Postgres database instance, but there was no error")
	}
}

func TestCreateDriverUsingDBInstance(t *testing.T) {
	postgresHost := os.Getenv("POSTGRES_HOST")

	database := "migrationtest"

	// prepare clean database
	connection, err := sql.Open("postgres", "postgres://postgres:@"+postgresHost+"/?sslmode=disable")

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := connection.Close()
		if err != nil {
			t.Errorf("unexpected error while closing the postgres connection: %v", err)
		}
	}()

	_, err = connection.Exec("CREATE DATABASE " + database)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_, err := connection.Exec("DROP DATABASE IF EXISTS " + database)
		if err != nil {
			t.Errorf("unexpected error while dropping the postgres database %s: %v", database, err)
		}
	}()

	db, err := sql.Open("postgres", "postgres://postgres:@"+postgresHost+"/"+database+"?sslmode=disable")

	if err != nil {
		t.Fatalf("could not open Postgres connection: %s", err)
	}

	driver, err := NewFromDB(db)

	if err != nil {
		t.Errorf("unable to create postgres driver: %s", err)
	}

	defer func() {
		err := driver.Close()
		if err != nil {
			t.Errorf("unexpected error %v while closing the postgres driver.", err)
		}
	}()
}
