// Package driver holds the driver interface.
package migration

// Driver is the interface type that needs to implemented by all drivers.
type Driver interface {

	// Close is the last function to be called.
	// Close any open connection here.
	Close() error

	// Migrate is the heart of the driver.
	// It will receive a file which the driver should apply
	// to its backend or whatever. The migration function should use
	// the pipe channel to return any errors or other useful information.
	Migrate(migration *PlannedMigration) error

	// Version returns all applied migration versions
	Versions() ([]string, error)
}
