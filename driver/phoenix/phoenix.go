package phoenix

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/Boostport/avatica"
	m "github.com/Boostport/migration"
)

type Phoenix struct {
	db *sql.DB
}

const phoenixTableName = "schema_migration"

// NewPhoenix creates a new Apache Phoenix driver.
// The DSN is documented here: https://github.com/Boostport/avatica#dsn-data-source-name
func NewPhoenix(dsn string) (m.Driver, error) {

	db, err := sql.Open("avatica", dsn)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	p := &Phoenix{
		db: db,
	}

	if err := p.ensureVersionTableExists(); err != nil {
		return nil, err
	}

	return p, nil
}

// Close closes the connection to the Apache Phoenix server.
func (driver *Phoenix) Close() error {
	err := driver.db.Close()
	return err
}

func (driver *Phoenix) ensureVersionTableExists() error {
	_, err := driver.db.Exec("CREATE TABLE IF NOT EXISTS " + phoenixTableName + " (version varchar not null primary key) TRANSACTIONAL=true")
	return err
}

// Migrate runs a migration.
func (driver *Phoenix) Migrate(migration *m.PlannedMigration) error {

	// TODO: Phoenix does not support DDL statements yet :( See PHOENIX-3358

	var content string

	if migration.Direction == m.Up {

		content = migration.Up

	} else if migration.Direction == m.Down {

		content = migration.Down
	}

	// Phoenix does not support multiple statements, so we need to do this.
	sqlStmts := strings.Split(content, ";")

	for _, sqlStmt := range sqlStmts {

		sqlStmt = strings.TrimSpace(sqlStmt)

		if len(sqlStmt) > 0 {

			if _, err := driver.db.Exec(sqlStmt); err != nil {
				return fmt.Errorf("Error executing statement: %s\n%s", err, sqlStmt)
			}
		}
	}

	if migration.Direction == m.Up {
		if _, err := driver.db.Exec("UPSERT INTO "+phoenixTableName+" (version) VALUES (?)", migration.ID); err != nil {
			return err
		}
	} else {
		if _, err := driver.db.Exec("DELETE FROM "+phoenixTableName+" WHERE version=?", migration.ID); err != nil {
			return err
		}
	}

	return nil
}

// Versions lists all the applied versions.
func (driver *Phoenix) Versions() ([]string, error) {
	versions := []string{}

	rows, err := driver.db.Query("SELECT version FROM " + phoenixTableName + " ORDER BY version DESC")

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
