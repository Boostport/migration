package postgres

import (
	"database/sql"
	"fmt"

	m "github.com/Boostport/migration"
	_ "github.com/lib/pq"
)

type Driver struct {
	db              *sql.DB
	useTransactions bool
}

const postgresTableName = "schema_migration"

// NewPostgres creates a new Driver driver.
// The DSN is documented here: https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
func New(dsn string, useTransactions bool) (m.Driver, error) {

	db, err := sql.Open("postgres", dsn)

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

// Close closes the connection to the Driver server.
func (driver *Driver) Close() error {
	err := driver.db.Close()
	return err
}

func (driver *Driver) ensureVersionTableExists() error {
	_, err := driver.db.Exec("CREATE TABLE IF NOT EXISTS " + postgresTableName + " (version varchar(255) not null primary key)")
	return err
}

// Migrate runs a migration.
func (driver *Driver) Migrate(migration *m.PlannedMigration) (err error) {

	var (
		content       string
		insertVersion string
	)

	if migration.Direction == m.Up {

		content = migration.Up
		insertVersion = "INSERT INTO " + postgresTableName + " (version) VALUES ($1)"

	} else if migration.Direction == m.Down {

		content = migration.Down
		insertVersion = "DELETE FROM " + postgresTableName + " WHERE version=$1"
	}

	if driver.useTransactions {

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

		if _, err = tx.Exec(insertVersion, migration.ID); err != nil {
			return fmt.Errorf("Error updating migration versions: %s", err)
		}
	} else {

		if _, err := driver.db.Exec(content); err != nil {
			return fmt.Errorf("Error executing statement: %s\n%s", err, content)
		}

		if _, err = driver.db.Exec(insertVersion, migration.ID); err != nil {
			return fmt.Errorf("Error updating migration versions: %s", err)
		}
	}

	return
}

// Versions lists all the applied versions.
func (driver *Driver) Versions() ([]string, error) {
	versions := []string{}

	rows, err := driver.db.Query("SELECT version FROM " + postgresTableName + " ORDER BY version DESC")

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
