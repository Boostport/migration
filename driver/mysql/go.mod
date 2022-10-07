module github.com/Boostport/migration/driver/mysql

go 1.18

require (
	github.com/Boostport/migration v1.1.2
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/go-sql-driver/mysql v1.6.0
)

replace github.com/Boostport/migration => ../..
