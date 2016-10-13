package mysql

import (
	"database/sql"
	"fmt"
	m "github.com/Boostport/migration"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

type MySQL struct {
	db *sql.DB
}

const mysqlTableName = "schema_migration"

// NewMySQL creates a new MySQL driver.
// The DSN is documented here: https://github.com/go-sql-driver/mysql#dsn-data-source-name
func NewMySQL(dsn string) (m.Driver, error) {

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	d := &MySQL{
		db: db,
	}

	if err := d.ensureVersionTableExists(); err != nil {
		return nil, err
	}

	return d, nil
}

// Close closes the connection to the MySQL server.
func (driver *MySQL) Close() error {

	if err := driver.db.Close(); err != nil {
		return err
	}

	return nil
}

func (driver *MySQL) ensureVersionTableExists() error {
	_, err := driver.db.Exec("CREATE TABLE IF NOT EXISTS " + mysqlTableName + " (version varchar(255) not null primary key)")
	return err
}

// Migrate runs a migration.
func (driver *MySQL) Migrate(migration *m.PlannedMigration) error {

	// Note: MySQL does not support DDL statements in a transaction. If DDL statements are
	// executed in a transaction, it is an implicit commit.
	// See: http://dev.mysql.com/doc/refman/5.7/en/implicit-commit.html
	var content string

	if migration.Direction == m.Up {

		content = migration.Up

	} else if migration.Direction == m.Down {

		content = migration.Down
	}

	// MySQL does not support multiple statements, so we need to do this.
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
		if _, err := driver.db.Exec("INSERT INTO "+mysqlTableName+" (version) VALUES (?)", migration.ID); err != nil {
			return err
		}
	} else {
		if _, err := driver.db.Exec("DELETE FROM "+mysqlTableName+" WHERE version=?", migration.ID); err != nil {
			return err
		}
	}

	return nil
}

// Versions lists all the applied versions.
func (driver *MySQL) Versions() ([]string, error) {
	versions := []string{}

	rows, err := driver.db.Query("SELECT version FROM " + mysqlTableName + " ORDER BY version DESC")

	if err != nil {
		return versions, err
	}

	defer rows.Close()

	for rows.Next() {
		var version string

		err := rows.Scan(&version)

		if err != nil {
			return versions, err
		}

		versions = append(versions, version)
	}

	err = rows.Err()

	return versions, err
}
