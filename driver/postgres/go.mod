module github.com/Boostport/migration/driver/postgres

go 1.18

require (
	github.com/Boostport/migration v1.0.0
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/jackc/pgconn v1.11.0
	github.com/jackc/pgx/v4 v4.15.0
)

replace github.com/Boostport/migration => ../..

require (
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.2.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.10.0 // indirect
	github.com/lib/pq v1.10.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
