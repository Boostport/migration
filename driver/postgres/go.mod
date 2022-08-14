module github.com/Boostport/migration/driver/postgres

go 1.18

require (
	github.com/Boostport/migration v1.1.1
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/jackc/pgconn v1.13.0
	github.com/jackc/pgx/v4 v4.17.0
)

replace github.com/Boostport/migration => ../..

require (
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.12.0 // indirect
	github.com/lib/pq v1.10.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	golang.org/x/text v0.3.7 // indirect
)
