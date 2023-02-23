module github.com/Boostport/migration/driver/postgres

go 1.18

require (
	github.com/Boostport/migration v1.1.2
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/jackc/pgconn v1.13.0
	github.com/jackc/pgx/v5 v5.0.1
)

replace github.com/Boostport/migration => ../..

require (
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90 // indirect
	golang.org/x/text v0.3.8 // indirect
)
