package postgres

import (
	"database/sql"
	"fmt"

	m "github.com/Boostport/migration"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

const postgresTableName = "schema_migration"

// NewPostgres creates a new Postgres driver.
// The DSN is documented here: https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
func NewPostgres(dsn string) (m.Driver, error) {

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	d := &Postgres{
		db: db,
	}

	if err := d.ensureVersionTableExists(); err != nil {
		return nil, err
	}

	return d, nil
}

// Close closes the connection to the Postgres server.
func (driver *Postgres) Close() error {
	err := driver.db.Close()
	return err
}

func (driver *Postgres) ensureVersionTableExists() error {
	_, err := driver.db.Exec("CREATE TABLE IF NOT EXISTS " + postgresTableName + " (version varchar(255) not null primary key)")
	return err
}

// Migrate runs a migration.
func (driver *Postgres) Migrate(migration *m.PlannedMigration) error {

	// Note: MySQL does not support DDL statements in a transaction. If DDL statements are
	// executed in a transaction, it is an implicit commit.
	// See: http://dev.mysql.com/doc/refman/5.7/en/implicit-commit.html
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

	if _, err = tx.Exec(content); err != nil {

		if err = tx.Rollback(); err != nil {
			return err
		}

		return fmt.Errorf("Error executing statement: %s\n%s", err, content)
	}

	if migration.Direction == m.Up {
		if _, err = tx.Exec("INSERT INTO "+postgresTableName+" (version) VALUES ($1)", migration.ID); err != nil {

			err = tx.Rollback()
			return err

		}
	} else {
		if _, err = tx.Exec("DELETE FROM "+postgresTableName+" WHERE version=$1", migration.ID); err != nil {

			err = tx.Rollback()
			return err

		}
	}

	err = tx.Commit()
	return err
}

// Versions lists all the applied versions.
func (driver *Postgres) Versions() ([]string, error) {
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
