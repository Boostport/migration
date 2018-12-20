package migration

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
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
