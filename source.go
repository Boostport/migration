package migration

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// Source is an interface that defines how a source can find and read migration files.
type Source interface {
	ListMigrationFiles() ([]string, error)
	GetMigrationFile(file string) (io.Reader, error)
}

// GolangMigrationSource implements migration.Source
type GolangMigrationSource struct {
	sync.Mutex
	migrations map[string]func() error
}

// NewGolangMigrationSource creates a source for storing Go functions as migrations.
func NewGolangMigrationSource() *GolangMigrationSource {
	return &GolangMigrationSource{
		migrations: map[string]func() error{},
	}
}

// AddMigration adds a new migration to the source. The file parameter follows the same conventions as you would use
// for a physical file for other types of migrations, however you should omit the file extension. Example: 1_init.up
// and 1_init.down
func (s *GolangMigrationSource) AddMigration(file string, direction Direction, migration func() error) {
	s.Lock()
	defer s.Unlock()

	if direction == Up {
		file += ".up"
	} else if direction == Down {
		file += ".down"
	}

	s.migrations[file+".go"] = migration
}

// GetMigration gets a golang migration
func (s *GolangMigrationSource) GetMigration(file string) func() error {
	s.Lock()
	defer s.Unlock()

	return s.migrations[file+".go"]
}

// ListMigrationFiles lists the available migrations in the source
func (s *GolangMigrationSource) ListMigrationFiles() ([]string, error) {
	var keys []string

	s.Lock()
	defer s.Unlock()

	for key := range s.migrations {
		keys = append(keys, key)
	}

	return keys, nil
}

// GetMigrationFile retrieves a migration given the filename.
func (s *GolangMigrationSource) GetMigrationFile(file string) (io.Reader, error) {
	s.Lock()
	defer s.Unlock()

	_, ok := s.migrations[file]
	if !ok {
		return nil, fmt.Errorf("migration %s does not exist", file)
	}

	return strings.NewReader(""), nil
}

// MemoryMigrationSource is a MigrationSource that uses migration sources in memory. It is mainly
// used for testing.
type MemoryMigrationSource struct {
	Files map[string]string
}

// ListMigrationFiles returns a list of memory migration files
func (m MemoryMigrationSource) ListMigrationFiles() ([]string, error) {
	var files []string

	for file := range m.Files {
		files = append(files, file)
	}

	return files, nil
}

// GetMigrationFile gets a memory migration file
func (m MemoryMigrationSource) GetMigrationFile(name string) (io.Reader, error) {
	content, ok := m.Files[name]
	if !ok {
		return nil, fmt.Errorf("the migration file %s does not exist", name)
	}

	return strings.NewReader(content), nil
}
