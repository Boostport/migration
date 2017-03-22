package golang

import (
	"fmt"
	"io"
	"strings"
	"sync"

	m "github.com/Boostport/migration"
)

type Config struct {
	sync.Mutex
	config map[interface{}]interface{}
}

// NewGolangConfig creates a concurrency-safe configuration map for passing configuration data and things like
// database handlers to your migrations.
func NewConfig() *Config {
	return &Config{
		config: map[interface{}]interface{}{},
	}
}

// Set adds a new key-value pair to the configuration.
func (c *Config) Set(key, val interface{}) {
	c.Lock()
	defer c.Unlock()

	c.config[key] = val
}

// Get retrieves a value from the configuration using the key.
func (c *Config) Get(key interface{}) interface{} {
	c.Lock()
	defer c.Unlock()

	return c.config[key]
}

type Source struct {
	sync.Mutex
	migrations map[string]func(c *Config) error
}

// NewGolangSource creates a source for storing Go functions as migrations.
func NewSource() *Source {
	return &Source{
		migrations: map[string]func(c *Config) error{},
	}
}

// AddMigration adds a new migration to the source. The file parameter follows the same conventions as you would use
// for a physical file for other types of migrations, however you should omit the file extension. Example: 1_init.up
// and 1_init.down
func (s *Source) AddMigration(file string, direction m.Direction, migration func(c *Config) error) {
	s.Lock()
	defer s.Unlock()

	if direction == m.Up {
		file += ".up"
	} else if direction == m.Down {
		file += ".down"
	}

	s.migrations[file+".go"] = migration
}

func (s *Source) getMigration(file string) func(c *Config) error {
	s.Lock()
	defer s.Unlock()

	return s.migrations[file+".go"]
}

// ListMigrationFiles lists the available migrations in the source
func (s *Source) ListMigrationFiles() ([]string, error) {
	keys := []string{}

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
		return nil, fmt.Errorf("Migration %s does not exist", file)
	}

	return strings.NewReader(""), nil
}

type Driver struct {
	source        *Source
	config        *Config
	updateVersion UpdateVersion
	applied       AppliedVersions
}

type UpdateVersion func(id string, direction m.Direction, config *Config) error

type AppliedVersions func(config *Config) ([]string, error)

// NewGolang creates a new Go migration driver. It requires a source a function for saving the executed migration version, a function for deleting a version
// that was migrated downwards, a function for listing all applied migrations and optionally a configuration.
func New(source *Source, updateVersion UpdateVersion, applied AppliedVersions, config *Config) (m.Driver, error) {
	return &Driver{
		source:        source,
		config:        config,
		updateVersion: updateVersion,
		applied:       applied,
	}, nil
}

func (g *Driver) Close() error {
	return nil
}

func (g *Driver) Migrate(migration *m.PlannedMigration) error {

	file := migration.ID

	if migration.Direction == m.Up {
		file += ".up"
	} else if migration.Direction == m.Down {
		file += ".down"
	}

	migrationFunc := g.source.getMigration(file)

	err := migrationFunc(g.config)

	if err != nil {
		return fmt.Errorf("Error executing golang migration: %s", err)
	}

	err = g.updateVersion(migration.ID, migration.Direction, g.config)

	if err != nil {
		return fmt.Errorf("Error executing golang update function: %s", err)
	}

	return nil
}

// Version returns all applied migration versions
func (g *Driver) Versions() ([]string, error) {
	return g.applied(g.config)
}
