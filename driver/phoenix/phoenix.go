package phoenix

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	m "github.com/Boostport/migration"
	"github.com/Boostport/migration/parser"
	avatica "github.com/apache/calcite-avatica-go/v5"
)

// Driver is the phoenix migration.Driver implementation
type Driver struct {
	db *sql.DB
}

const phoenixTableName = "schema_migration"

// New creates a new Apache Avatica Driver.
// The DSN is documented here: https://calcite.apache.org/avatica/docs/go_client_reference.html#dsn-data-source-name
func New(dsn string) (m.Driver, error) {
	db, err := sql.Open("avatica", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	p := &Driver{
		db: db,
	}
	if err := p.ensureVersionTableExists(); err != nil {
		return nil, err
	}

	return p, nil
}

// NewFromDB returns a phoenix driver from a sql.DB
func NewFromDB(db *sql.DB) (m.Driver, error) {
	if _, ok := db.Driver().(*avatica.Driver); !ok {
		return nil, errors.New("database instance is not using the avatica driver")
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

// Close closes the connection to the Apache Driver server.
func (driver *Driver) Close() error {
	err := driver.db.Close()
	return err
}

func (driver *Driver) ensureVersionTableExists() error {
	_, err := driver.db.Exec("CREATE TABLE IF NOT EXISTS " + phoenixTableName + " (version varchar not null primary key) TRANSACTIONAL=true")
	return err
}

// Migrate runs a migration.
func (driver *Driver) Migrate(migration *m.PlannedMigration) error {
	// TODO: Driver does not support DDL statements in a transaction yet :( See PHOENIX-3358
	var migrationStatements *parser.ParsedMigration

	if migration.Direction == m.Up {
		migrationStatements = migration.Up
	} else if migration.Direction == m.Down {
		migrationStatements = migration.Down
	}

	for _, sqlStmt := range migrationStatements.Statements {
		// Special case for Phoenix. We force a statement split here, because Phoenix SQL statements must not be terminated with ;.
		// In addition, this explicitly splits the SQL statements into its constituent statements.
		splitted := strings.Split(sqlStmt, ";")

		for _, content := range splitted {
			if len(strings.TrimSpace(content)) > 0 {
				if _, err := driver.db.Exec(content); err != nil {
					return fmt.Errorf("Error executing statement: %s\n%s", err, content)
				}
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
func (driver *Driver) Versions() ([]string, error) {
	var versions []string

	rows, err := driver.db.Query("SELECT version FROM " + phoenixTableName + " ORDER BY version DESC")
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
