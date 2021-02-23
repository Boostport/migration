// +build go1.16

package migration

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path"
)

// EmbedMigrationSource uses a embed.FS that is used to embed files natively in Go 1.16+
type EmbedMigrationSource struct {
	EmbedFS embed.FS

	// The path in the embed FS to use
	Dir string
}

// ListMigrationFiles returns a list of embedded migration files
func (e EmbedMigrationSource) ListMigrationFiles() ([]string, error) {

	var f fs.FS = e.EmbedFS

	if e.Dir != "" {

		var err error

		f, err = fs.Sub(f, e.Dir)

		if err != nil {
			return nil, fmt.Errorf("error opening subdirectory in embed fs: %s", err)
		}

	}

	files, err := fs.ReadDir(f, ".")

	if err != nil {
		return nil, fmt.Errorf("error reading directory from embed fs: %w", err)
	}

	var migrations []string

	for _, file := range files {

		if file.IsDir() {
			continue
		}

		migrations = append(migrations, file.Name())
	}

	return migrations, nil
}

// GetMigrationFile gets an embedded migration file
func (e EmbedMigrationSource) GetMigrationFile(name string) (io.Reader, error) {
	file, err := fs.ReadFile(e.EmbedFS, path.Join(e.Dir, name))
	return bytes.NewReader(file), err
}
