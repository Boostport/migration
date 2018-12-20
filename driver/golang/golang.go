package golang

import (
	"fmt"
	"io"
	"strings"
	"sync"

	m "github.com/Boostport/migration"
)

// Source implements migration.Source
type Source struct {
	sync.Mutex
	migrations map[string]func() error
}

// NewSource creates a source for storing Go functions as migrations.
func NewSource() *Source {
	return &Source{
		migrations: map[string]func() error{},
	}
}

// AddMigration adds a new migration to the source. The file parameter follows the same conventions as you would use
// for a physical file for other types of migrations, however you should omit the file extension. Example: 1_init.up
// and 1_init.down
func (s *Source) AddMigration(file string, direction m.Direction, migration func() error) {
	s.Lock()
	defer s.Unlock()

	if direction == m.Up {
		file += ".up"
	} else if direction == m.Down {
		file += ".down"
	}

	s.migrations[file+".go"] = migration
}

func (s *Source) getMigration(file string) func() error {
	s.Lock()
	defer s.Unlock()

	return s.migrations[file+".go"]
}

// ListMigrationFiles lists the available migrations in the source
func (s *Source) ListMigrationFiles() ([]string, error) {
	var keys []string

	s.Lock()
	defer s.Unlock()

	for key := range s.migrations {
		keys = append(keys, key)
	}

	return keys, nil
}

// GetMigrationFile retrieves a migration given the filename.
func (s *Source) GetMigrationFile(file string) (io.Reader, error) {

	s.Lock()
	defer s.Unlock()

	_, ok := s.migrations[file]

	if !ok {
		return nil, fmt.Errorf("migration %s does not exist", file)
	}

	return strings.NewReader(""), nil
}

// Driver is the golang migration.Driver implementation
type Driver struct {
	source        *Source
	updateVersion UpdateVersion
	applied       AppliedVersions
}

// UpdateVersion takes an id and a direction and returns an error if something fails
type UpdateVersion func(id string, direction m.Direction) error

// AppliedVersions returns a list of applied versions and an error if something fails
type AppliedVersions func() ([]string, error)

// New creates a new Go migration driver. It requires a source a function for saving the executed migration version, a function for deleting a version
// that was migrated downwards, a function for listing all applied migrations and optionally a configuration.
func New(source *Source, updateVersion UpdateVersion, applied AppliedVersions) (m.Driver, error) {
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

	migrationFunc := g.source.getMigration(file)

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
