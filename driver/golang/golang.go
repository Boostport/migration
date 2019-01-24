package golang

import (
	"fmt"

	m "github.com/Boostport/migration"
)

// Driver is the golang migration.Driver implementation
type Driver struct {
	source        *m.GolangMigrationSource
	updateVersion UpdateVersion
	applied       AppliedVersions
}

// UpdateVersion takes an id and a direction and returns an error if something fails
type UpdateVersion func(id string, direction m.Direction) error

// AppliedVersions returns a list of applied versions and an error if something fails
type AppliedVersions func() ([]string, error)

// New creates a new Go migration driver. It requires a source a function for saving the executed migration version, a function for deleting a version
// that was migrated downwards, a function for listing all applied migrations and optionally a configuration.
func New(source *m.GolangMigrationSource, updateVersion UpdateVersion, applied AppliedVersions) (m.Driver, error) {
	return &Driver{
		source:        source,
		updateVersion: updateVersion,
		applied:       applied,
	}, nil
}

// Close is the migration.Driver implementation of io.Closer
func (g *Driver) Close() error {
	return nil
}

// Migrate executes a planned migration
func (g *Driver) Migrate(migration *m.PlannedMigration) error {
	file := migration.ID

	if migration.Direction == m.Up {
		file += ".up"
	} else if migration.Direction == m.Down {
		file += ".down"
	}

	migrationFunc := g.source.GetMigration(file)

	err := migrationFunc()

	if err != nil {
		return fmt.Errorf("error executing golang migration: %s", err)
	}

	err = g.updateVersion(migration.ID, migration.Direction)

	if err != nil {
		return fmt.Errorf("error executing golang update function: %s", err)
	}

	return nil
}

// Versions returns all applied migration versions
func (g *Driver) Versions() ([]string, error) {
	return g.applied()
}
