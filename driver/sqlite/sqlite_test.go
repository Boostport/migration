package sqlite

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/Boostport/migration"
	"github.com/Boostport/migration/parser"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
)

func TestSQLiteDriver(t *testing.T) {
	for _, useTransactions := range []bool{true, false} {
		driver, err := New("file::memory:?cache=shared&_busy_timeout=50000", useTransactions)

		if err != nil {
			t.Errorf("unable to open connection to server: %s", err)
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

		_, err = driver.(*Driver).db.Exec("INSERT INTO test_table1 (id) values (1)")
		if err != nil {
			t.Errorf("unexpected error while testing if migration succeeded: %s", err)
		}

		_, err = driver.(*Driver).db.Exec("INSERT INTO test_table2 (id) values (1)")
		if err != nil {
			t.Errorf("unexpected error while testing if migration succeeded: %s", err)
		}

		err = driver.Migrate(migrations[1])
		if err != nil {
			t.Errorf("unexpected error while running migration: %s", err)
		}

		if _, err := driver.(*Driver).db.Exec("INSERT INTO test_table2 (id) values (1)"); err != nil {
			reg := regexp.MustCompile(`^no such table: .+`)

			if !reg.MatchString(err.Error()) {
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
			t.Errorf("expected %d versions to be applied, %d was actually applied", 2, len(versions))
		}

		_, err = driver.(*Driver).db.Exec("DROP TABLE IF EXISTS " + sqliteTableName)
		if err != nil {
			t.Errorf("unexpected error %v while droping the table: %s", err, sqliteTableName)
		}

		err = driver.Close()
		if err != nil {
			t.Errorf("unexpected error %v while closing the sqlite driver", err)
		}
	}
}

func TestCreateDriverUsingInvalidDBInstance(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening stub database connection: %s", err)
	}

	_, err = NewFromDB(db)
	if err == nil {
		t.Error("expected error when creating SQLite driver with a non-SQLite database instance, but there was no error")
	}
}

func TestCreateDriverUsingDBInstance(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared&_busy_timeout=50000")
	if err != nil {
		t.Fatal(err)
	}

	driver, err := NewFromDB(db)
	if err != nil {
		t.Errorf("unable to create SQLite driver: %s", err)
	}

	defer func() {
		err := driver.Close()
		if err != nil {
			t.Errorf("unexpected error %v while closing the sqlite driver", err)
		}
	}()
}
