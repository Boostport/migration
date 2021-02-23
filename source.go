package migration

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/markbates/pkger"
)

// Source is an interface that defines how a source can find and read migration files.
type Source interface {
	ListMigrationFiles() ([]string, error)
	GetMigrationFile(file string) (io.Reader, error)
}

// GoBindataMigrationSource is a MigrationSource that uses migration files embedded in a Go application using go-bindata.
type GoBindataMigrationSource struct {
	// Asset should return content of file in path if exists
	Asset func(path string) ([]byte, error)

	// AssetDir should return list of files in the path
	AssetDir func(path string) ([]string, error)

	// Path in the bindata to use.
	Dir string
}

// ListMigrationFiles returns a list of gobindata migration files
func (a GoBindataMigrationSource) ListMigrationFiles() ([]string, error) {
	return a.AssetDir(a.Dir)
}

// GetMigrationFile gets a gobindata migration file
func (a GoBindataMigrationSource) GetMigrationFile(name string) (io.Reader, error) {
	file, err := a.Asset(path.Join(a.Dir, name))
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(file), nil
}

// PackrBox avoids pulling in the packr library for everyone, mimics the bits of
// packr.Box that we need.
type PackrBox interface {
	List() []string
	Find(name string) ([]byte, error)
}

// PackrMigrationSource holds the box and dir info
type PackrMigrationSource struct {
	Box PackrBox

	// The path in the packr box to use
	Dir string
}

// ListMigrationFiles returns a list of packr migration files
func (p PackrMigrationSource) ListMigrationFiles() ([]string, error) {
	files := p.Box.List()
	var migrations []string
	prefix := ""

	dir := path.Clean(p.Dir)
	if dir != "." {
		prefix = fmt.Sprintf("%s/", dir)
	}

	for _, file := range files {
		if !strings.HasPrefix(file, prefix) {
			continue
		}
		name := strings.TrimPrefix(file, prefix)
		if strings.Contains(name, "/") {
			continue
		}

		migrations = append(migrations, name)
	}

	return migrations, nil
}

// GetMigrationFile gets a packr migration file
func (p PackrMigrationSource) GetMigrationFile(name string) (io.Reader, error) {
	file, err := p.Box.Find(path.Join(p.Dir, name))

	return bytes.NewReader(file), err
}

// PkgerMigrationSource holds the underlying pkger and dir info
type PkgerMigrationSource struct {
	// The path to use
	Dir string
}

// ListMigrationFiles returns a list of pkger migration files
func (p PkgerMigrationSource) ListMigrationFiles() ([]string, error) {

	var migrations []string

	err := pkger.Walk(p.Dir, func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return nil
		}

		migrations = append(migrations, info.Name())

		return nil
	})

	if err != nil {
		return migrations, fmt.Errorf("error listing migration files: %s", err)
	}

	return migrations, nil
}

// GetMigrationFile gets a pkger migration file
func (p PkgerMigrationSource) GetMigrationFile(name string) (io.Reader, error) {
	return pkger.Open(path.Join(p.Dir, name))
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
