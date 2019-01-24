package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	m "github.com/Boostport/migration"
	"github.com/Boostport/migration/parser"
	sqlite3 "github.com/mattn/go-sqlite3"
)

// Driver is the sqlite migration.Driver implementation
type Driver struct {
	db              *sql.DB
	useTransactions bool
}

const sqliteTableName = "schema_migration"

// New creates a new Driver driver.
// The DSN is documented here: https://godoc.org/github.com/mattn/go-sqlite3#SQLiteDriver.Open
func New(dsn string, useTransactions bool) (m.Driver, error) {

	db, err := sql.Open("sqlite3", dsn)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	d := &Driver{
		db:              db,
		useTransactions: useTransactions,
	}

	if err := d.ensureVersionTableExists(); err != nil {
		return nil, err
	}

	return d, nil
}

// NewFromDB returns a sqlite driver from a sql.db
func NewFromDB(db *sql.DB) (m.Driver, error) {
	if _, ok := db.Driver().(*sqlite3.SQLiteDriver); !ok {
		return nil, errors.New("database instance is not using the postgres driver")
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	d := &Driver{
		db: db,
	}

	if err := d.ensureVersionTableExists(); err != nil {
		return nil, err
	}

	return d, nil
}

// Close closes the connection to the Driver server.
func (driver *Driver) Close() error {
	err := driver.db.Close()
	return err
}

func (driver *Driver) ensureVersionTableExists() error {
	_, err := driver.db.Exec("CREATE TABLE IF NOT EXISTS " + sqliteTableName + " (version varchar(255) not null primary key)")
	return err
}

// Migrate runs a migration.
func (driver *Driver) Migrate(migration *m.PlannedMigration) (err error) {
	var (
		migrationStatements *parser.ParsedMigration
		insertVersion       string
	)

	if migration.Direction == m.Up {
		migrationStatements = migration.Up
		insertVersion = "INSERT INTO " + sqliteTableName + " (version) VALUES (?)"

	} else if migration.Direction == m.Down {
		migrationStatements = migration.Down
		insertVersion = "DELETE FROM " + sqliteTableName + " WHERE version=?"
	}

	if driver.useTransactions {
		tx, err := driver.db.Begin()
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				if errRb := tx.Rollback(); errRb != nil {
					err = fmt.Errorf("error rolling back: %s\n%s", errRb, err)
				}
				return
			}
			err = tx.Commit()
		}()

		for _, statement := range migrationStatements.Statements {
			if _, err = tx.Exec(statement); err != nil {
				return fmt.Errorf("error executing statement: %s\n%s", err, statement)
			}
		}

		if _, err = tx.Exec(insertVersion, migration.ID); err != nil {
			return fmt.Errorf("error updating migration versions: %s", err)
		}
	} else {
		for _, statement := range migrationStatements.Statements {
			if _, err := driver.db.Exec(statement); err != nil {
				return fmt.Errorf("error executing statement: %s\n%s", err, statement)
			}
		}
		if _, err = driver.db.Exec(insertVersion, migration.ID); err != nil {
			return fmt.Errorf("error updating migration versions: %s", err)
		}
	}

	return
}

// Versions lists all the applied versions.
func (driver *Driver) Versions() ([]string, error) {
	var versions []string

	rows, err := driver.db.Query("SELECT version FROM " + sqliteTableName + " ORDER BY version DESC")
	if err != nil {
		return versions, err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var version string

		err = rows.Scan(&version)

		if err != nil {
			return versions, err
		}

		versions = append(versions, version)
	}

	err = rows.Err()

	return versions, err
}
