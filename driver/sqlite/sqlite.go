package sqlite

import (
	"database/sql"
	"fmt"

	m "github.com/Boostport/migration"
	_ "github.com/mattn/go-sqlite3"
)

type Driver struct {
	db *sql.DB
}

const sqliteTableName = "schema_migration"

// NewSQLite creates a new Driver driver.
// The DSN is documented here: https://godoc.org/github.com/mattn/go-sqlite3#SQLiteDriver.Open
func New(dsn string) (m.Driver, error) {

	db, err := sql.Open("sqlite3", dsn)

	if err != nil {
		return nil, err
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

	var content string

	if migration.Direction == m.Up {

		content = migration.Up

	} else if migration.Direction == m.Down {

		content = migration.Down
	}

	tx, err := driver.db.Begin()

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if errRb := tx.Rollback(); errRb != nil {
				err = fmt.Errorf("Error rolling back: %s\n%s", errRb, err)
			}
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec(content); err != nil {

		return fmt.Errorf("Error executing statement: %s\n%s", err, content)
	}

	if migration.Direction == m.Up {
		if _, err = tx.Exec("INSERT INTO "+sqliteTableName+" (version) VALUES (?)", migration.ID); err != nil {

			return

		}
	} else {
		if _, err = tx.Exec("DELETE FROM "+sqliteTableName+" WHERE version=?", migration.ID); err != nil {

			return

		}
	}
	return
}

// Versions lists all the applied versions.
func (driver *Driver) Versions() ([]string, error) {
	versions := []string{}

	rows, err := driver.db.Query("SELECT version FROM " + sqliteTableName + " ORDER BY version DESC")

	if err != nil {
		return versions, err
	}

	defer rows.Close()

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
