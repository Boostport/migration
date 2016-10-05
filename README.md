# Migration
Simple and pragmatic migrations for Go applications.

## Why yet another migration library?
We wanted a migration library with the following features:
- Open to extension for all sorts of databases, not just `database/sql` drivers or an ORM.
- Easily embeddable in a Go application.
- Support for embedding migration files directly into the app.

We narrowed our focus down to 2 contenders: [sql-migrate](https://github.com/rubenv/sql-migrate)
and [migrate](https://github.com/mattes/migrate/)

`sql-migrate` leans heavily on the [gorp](https://github.com/go-gorp/gorp) ORM library to perform migrations.
Unfortunately, this means that we were restricted to databases supported by `gorp`. It is easily embeddable in a
Go app and supports embedding migration files directly into the Go binary. If database was a bit more flexible,
we would have gone with it.

`migrate` is highly extensible, and adding support for another database is extremely trivial. However, due to it using
the scheme in the dsn to determine which database driver to use, it prevented us from easily implementing an Apache
Phoenix driver, which uses the scheme to determine if we should connect over `http` or `https`. Due to the way the
project is structured, it was also almost impossible to add support for embeddable migration files without major
changes.