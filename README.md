# Migration
[![GoDoc](https://godoc.org/github.com/Boostport/migration?status.png)](https://godoc.org/github.com/Boostport/migration)
[![Tests Status](https://github.com/Boostport/migration/workflows/Tests/badge.svg)](https://github.com/Boostport/migration)
[![Test Coverage](https://api.codeclimate.com/v1/badges/caad2af95fa34fd23a2e/test_coverage)](https://codeclimate.com/github/Boostport/migration/test_coverage)

Simple and pragmatic migrations for Go applications.

## Features
- Super simple driver interface to allow easy implementation for more database/migration drivers.
- Embeddable migration files.
- Support for up/down migrations.
- Atomic migrations (where possible, depending on database support).
- Support for using Go code as migrations

## Drivers
- Apache Phoenix
- Golang (runs generic go functions)
- MySQL
- PostgreSQL
- SQLite

## Quickstart
```go
// Create migration source
//go:embed migrations
var embedFS embed.FS

embedSource := &migration.EmbedMigrationSource{
	EmbedFS: embedFS,
	Dir:     "migrations",
}

// Create driver
driver, err := mysql.New("root:@tcp(localhost)/mydatabase?multiStatements=true")

// Run all up migrations
applied, err := migration.Migrate(driver, embedSource, migration.Up, 0)

// Remove the last 2 migrations
applied, err := migration.Migrate(driver, embedSource, migration.Down, 2)
```

## Writing migrations
Migrations are extremely simple to write:
- Separate your up and down migrations into different files. For example, `1_init.up.sql` and `1_init.down.sql`.
- Prefix your migration with a number or timestamp for versioning: `1_init.up.sql` or `1475813115_init.up.sql`.
- The file-extension can be anything you want, but must be present. For example, `1_init.up.sql` is valid, but
`1_init.up` is not,
- Note: Underscores (`_`) must be used to separate the number and description in the filename.

Let's say we want to write our first migration to initialize the database.

In that case, we would have a file called `1_init.up.sql` containing SQL statements for the
up migration:

```sql
CREATE TABLE test_data (
  id BIGINT NOT NULL PRIMARY KEY,
)
```

We also create a `1_init.down.sql` file containing SQL statements for the down migration:
```sql
DROP TABLE IF EXISTS test_data
```

By default, migrations are run within a transaction. If you do not want a migration to run within a transaction,
start the migration file with `-- +migration NoTransaction`:

```sql
-- +migration NoTransaction

CREATE TABLE test_data1 (
  id BIGINT NOT NULL PRIMARY KEY,
)

CREATE TABLE test_data2 (
  id BIGINT NOT NULL PRIMARY KEY,
)
```

If you would like to create stored procedures, triggers or complex statements that contain semicolns, use `BeginStatement`
and `EndStatement` to delineate them:

```sql
CREATE TABLE test_data1 (
  id BIGINT NOT NULL PRIMARY KEY,
)

CREATE TABLE test_data2 (
  id BIGINT NOT NULL PRIMARY KEY,
)

-- +migration BeginStatement
CREATE TRIGGER`test_trigger_1`BEFORE UPDATE ON`test_data1`FOR EACH ROW BEGIN
		INSERT INTO test_data2
		SET id = OLD.id;
END
-- +migration EndStatement
```

## Embedding migration files

### Using [go:embed](https://golang.org/pkg/embed/) (Recommended for Go 1.16+)
This is the recommended method for embedding migration files if you are using Go 1.16+. The `go:embed` Go's built-in
method to embed files into the built binary and does not require any external tools.

Assuming your migration files are in `migrations/`, initialize a `EmbededSource`:
```go
//go:embed migrations
var embedFS embed.FS

assetMigration := &migration.EmbedSource{
    EmbedFS: embedFS,
    Dir:     "migrations",
}
```

### Using [pkger](https://github.com/markbates/pkger)
Assuming your migration files are in Assuming your migration files are in `migrations/`, initialize `pkger` and a `PkgerMigrationSource`:
```go
dir := pkger.Include("/test-migrations") // Remember to include forward slash at the beginning of the directory's name

pkgerSource := &migration.PkgerMigrationSource{
    Dir: dir,
}
```

During development, pkger will read the migration files from disk. When building for production, run `pkger` to generate
a Go file containing your migrations. For more information, see the [pkger documenation](https://github.com/markbates/pkger#usage).

### Using [packr](https://github.com/gobuffalo/packr)
Assuming your migration files are in `migrations/`, initialize a `PackrMigrationSource`:
```go
packrSource := &migration.PackrMigrationSource{
	Box: packr.New("migrations", "migrations"),
}
```

If your migrations are contained in a subdirectory inside your packr box, you can point to it using the `Dir` property:
```go
packrSource := &migration.PackrMigrationSource{
	Box: packr.New("migrations", "."),
	Dir: "migrations",
}
```

During development, packr will read the migration files from disk. When building for production, run `packr` to generate
a Go file containing your migrations, or use `packr build` to build for your binary. For more information, see the
[packr documenation](https://github.com/gobuffalo/packr#building-a-binary-the-easy-way).

### Using [go-bindata](https://github.com/go-bindata/go-bindata)
*Note: We recommend using packr as it allows you to use migrations from disk during development*

In the simplest case, assuming your migration files are in `migrations/`, just run:
```
go-bindata -o bindata.go -pkg myapp migrations/
```

Then, use `GoBindataMigrationSource` to find the migrations:
```go
goBindataSource := &migration.GoBindataMigrationSource{
    Asset:    Asset,
    AssetDir: AssetDir,
    Dir:      "test-migrations",
}
```

The `Asset` and `AssetDir` functions are generated by `go-bindata`.

## Using Go for migrations
Sometimes, we might be working with a database or have a situation where the query language is not expressive enough
to perform the required migrations. For example, we might have to get some data out of the database, perform some 
transformations and then write it back. For these type of situations, you can use Go for migrations.

When using Go for migrations, create a `golang.Source` using `golang.NewSource()`. Then, simply add migrations to the source
using the `AddMigration()` method. You will need to pass in the name of the migration without the extension and direction, e.g.
`1_init`. For the second parameter, pass in the direction (`migration.Up` or `migration.Down`) and for the third parameter,
pass in a function or method with this signature: `func() error` for running the migration.

Finally, you need to define 2 functions:
- A function for writing or deleting an applied migration matching this signature: `func(id string, direction migration.Direction) error`
- A function for getting a list of applied migrations matching this signature: `func() ([]string, error)`

These are required for initializing the driver:
```go
driver, err := golang.New(source, updateVersion, applied)
```

Here's a quick example:
```go
source := migration.NewGolangMigrationSource()

source.AddMigration("1_init", migration.Up, func() error {
    // Run up migration here
})

source.AddMigration("1_init", migration.Down, func() error {
    // Run down migration here
})

// Define functions
applied := func() ([]string, error) {
    // Return list of applied migrations
}

updateVersion := func(id string, direction migration.Direction) error {
    // Write or delete applied migration in storage
}

// Create driver
driver, err := golang.New(source, updateVersion, applied)

// Run migrations
count, err = migration.Migrate(driver, source, migration.Up, 0)
```

## TODO (Pull requests welcomed!)
- [ ] Command line program to run migrations
- [ ] More drivers

## Why yet another migration library?
We wanted a migration library with the following features:
- Open to extension for all sorts of databases, not just `database/sql` drivers or an ORM.
- Easily embeddable in a Go application.
- Support for embedding migration files directly into the app.

We narrowed our focus down to 2 contenders: [sql-migrate](https://github.com/rubenv/sql-migrate)
and [migrate](https://github.com/mattes/migrate/)

`sql-migrate` leans heavily on the [gorp](https://github.com/go-gorp/gorp) ORM library to perform migrations.
Unfortunately, this means that we were restricted to databases supported by `gorp`. It is easily embeddable in a
Go app and supports embedding migration files directly into the Go binary. If database support was a bit more flexible,
we would have gone with it.

`migrate` is highly extensible, and adding support for another database is extremely trivial. However, due to it using
the scheme in the dsn to determine which database driver to use, it prevented us from easily implementing an Apache
Phoenix driver, which uses the scheme to determine if we should connect over `http` or `https`. Due to the way the
project is structured, it was also almost impossible to add support for embeddable migration files without major
changes.

## Contributing
We automatically run some linters using [golangci-lint](https://github.com/golangci/golangci-lint) to check code quality
before merging it. This is executed using a [Makefile](Makefile) target. 

You should run and ensure all the checks pass locally before submitting a pull request. The version of
[golangci-lint](https://github.com/golangci/golangci-lint) to be used is pinned in `go.mod`.

To execute the linters:
1. Install `make`.
2. Install [golangci-lint](https://github.com/golangci/golangci-lint) by executing `go install github.com/golangci/golangci-lint/cmd/golangci-lint`.
3. Execute `make sanity-check`.

## License
This library is licensed under the Apache 2 License.
