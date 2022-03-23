package mysql

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/Boostport/migration"
	"github.com/Boostport/migration/parser"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
)

func TestMySQLDriver(t *testing.T) {
	mysqlHost := os.Getenv("MYSQL_HOST")
	database := "migrationtest"

	// prepare clean database
	connection, err := sql.Open("mysql", "root:@tcp("+mysqlHost+")/")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := connection.Close()
		if err != nil {
			t.Errorf("unexpected error while closing the mysql connection: %v", err)
		}
	}()

	for start := time.Now(); ; {

		err = connection.Ping()

		if err == nil {
			break
		}

		if time.Since(start) > 10*time.Second {
			t.Fatal("Timed out while waiting for MySQL server")
		}
	}

	_, err = connection.Exec("CREATE DATABASE IF NOT EXISTS " + database)
	if err != nil {
		t.Fatal(err)
	}

	_, err = connection.Exec("USE " + database)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := New("root:@tcp(" + mysqlHost + ")/" + database + "?multiStatements=true")
	if err != nil {
		t.Errorf("unable to open connection to mysql server: %s", err)
	}
	defer func() {
		err := driver.Close()
		if err != nil {
			t.Errorf("unexpected error while closing the mysql driver: %v", err)
		}
	}()

	defer func() {
		_, err := connection.Exec("DROP DATABASE IF EXISTS " + database)
		if err != nil {
			t.Errorf("unexpected error while dropping the mysql database %s: %v", database, err)
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
		t.Errorf("unexpected error while running migration: %s", err)
	}

	_, err = connection.Exec("INSERT INTO test_table1 (id) values (1)")
	if err != nil {
		t.Errorf("unexpected error while testing if migration succeeded: %s", err)
	}

	_, err = connection.Exec("INSERT INTO test_table2 (id) values (1)")
	if err != nil {
		t.Errorf("unexpected error while testing if migration succeeded: %s", err)
	}

	err = driver.Migrate(migrations[1])
	if err != nil {
		t.Errorf("unexpected error while running migration: %s", err)
	}

	if _, err = connection.Exec("INSERT INTO test_table2 (id) values (1)"); err != nil {
		if err.(*mysql.MySQLError).Number != 1146 {
			t.Errorf("received an error while inserting into a non-existent table, but it was not a table_undefined error: %s", err)
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
}

func TestCreateDriverUsingInvalidDBInstance(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening stub database connection: %s", err)
	}

	_, err = NewFromDB(db)
	if err == nil {
		t.Error("expected error when creating MySQL driver with a non-MySQL database instance, but there was no error")
	}
}

func TestCreateDriverUsingDBInstance(t *testing.T) {
	mysqlHost := os.Getenv("MYSQL_HOST")
	database := "migrationtest"

	// prepare clean database
	connection, err := sql.Open("mysql", "root:@tcp("+mysqlHost+")/")
	if err != nil {
		t.Fatal(err)
	}

	_, err = connection.Exec("CREATE DATABASE IF NOT EXISTS " + database)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := connection.Close()
		if err != nil {
			t.Errorf("unexpected error while closing the mysql connection: %v", err)
		}
	}()

	defer func() {
		_, err := connection.Exec("DROP DATABASE IF EXISTS " + database)
		if err != nil {
			t.Errorf("unexpected error while dropping the mysql database %s: %v", database, err)
		}
	}()

	db, err := sql.Open("mysql", "root:@tcp("+mysqlHost+")/"+database+"?multiStatements=true")
	if err != nil {
		t.Fatalf("could not open MySQL connection: %s", err)
	}

	driver, err := NewFromDB(db)
	if err != nil {
		t.Errorf("unable to create MySQL driver: %s", err)
	}

	defer func() {
		err := driver.Close()
		if err != nil {
			t.Errorf("unexpected error while closing the mysql driver: %v", err)
		}
	}()
}
