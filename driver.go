package migration

// Driver is the interface type that needs to implemented by all drivers.
type Driver interface {
	// Close is the last function to be called.
	// Close any open connection here.
	Close() error

	// Migrate is the heart of the driver.
	// It will receive a PlannedMigration which the driver should apply
	// to its backend or whatever.
	Migrate(migration *PlannedMigration) error

	// Version returns all applied migration versions
	Versions() ([]string, error)
}
