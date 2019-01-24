package parser

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {

	testMigrations := []struct {
		statements  string
		result      []string
		transaction bool
	}{
		{
			statements: `
				CREATE TABLE test_table1 (id integer not null primary key);

				CREATE TABLE test_table2 (id integer not null primary key);

				-- +migration BeginStatement
				CREATE TRIGGER ` + "`test_trigger`" + ` BEFORE UPDATE ON ` + "`test_table1`" + ` FOR EACH ROW BEGIN
				    INSERT INTO test_table2
				    SET id = OLD.id;
				END
				-- +migration EndStatement

				CREATE TABLE test_table3 (id integer not null primary key);
				`,
			result: []string{
				`
				CREATE TABLE test_table1 (id integer not null primary key);

				CREATE TABLE test_table2 (id integer not null primary key);

`,

				`				CREATE TRIGGER ` + "`test_trigger`" + ` BEFORE UPDATE ON ` + "`test_table1`" + ` FOR EACH ROW BEGIN
				    INSERT INTO test_table2
				    SET id = OLD.id;
				END
`,

				`
				CREATE TABLE test_table3 (id integer not null primary key);
				`,
			},

			transaction: true,
		},

		{
			statements: `
				CREATE TABLE test_table1 (id integer not null primary key);

				CREATE TABLE test_table2 (id integer not null primary key);

				-- +migration BeginStatement
				CREATE TRIGGER ` + "`test_trigger`" + ` BEFORE UPDATE ON ` + "`test_table1`" + ` FOR EACH ROW BEGIN
				    INSERT INTO test_table2
				    SET id = OLD.id;
				END
				-- +migration EndStatement
				`,

			result: []string{
				`
				CREATE TABLE test_table1 (id integer not null primary key);

				CREATE TABLE test_table2 (id integer not null primary key);

`,

				`				CREATE TRIGGER ` + "`test_trigger`" + ` BEFORE UPDATE ON ` + "`test_table1`" + ` FOR EACH ROW BEGIN
				    INSERT INTO test_table2
				    SET id = OLD.id;
				END
`,
			},

			transaction: true,
		},

		{
			statements: `-- +migration NoTransaction

				CREATE TABLE test_table1 (id integer not null primary key);

				CREATE TABLE test_table2 (id integer not null primary key);

				-- +migration BeginStatement
				CREATE TRIGGER ` + "`test_trigger`" + ` BEFORE UPDATE ON ` + "`test_table1`" + ` FOR EACH ROW BEGIN
				    INSERT INTO test_table2
				    SET id = OLD.id;
				END
				-- +migration EndStatement

				CREATE TABLE test_table3 (id integer not null primary key);
				`,

			result: []string{
				`
				CREATE TABLE test_table1 (id integer not null primary key);`,

				`

				CREATE TABLE test_table2 (id integer not null primary key);

`,

				`				CREATE TRIGGER ` + "`test_trigger`" + ` BEFORE UPDATE ON ` + "`test_table1`" + ` FOR EACH ROW BEGIN
				    INSERT INTO test_table2
				    SET id = OLD.id;
				END
`,

				`
				CREATE TABLE test_table3 (id integer not null primary key);
				`,
			},

			transaction: false,
		},
	}

	for i, testCase := range testMigrations {
		parsed, err := Parse(strings.NewReader(testCase.statements))
		if err != nil {
			t.Errorf("Unexpected error while parsing statements for test case %d: %s", i, err)
		}
		if parsed.UseTransaction != testCase.transaction {
			t.Errorf("Transactions for test case %d are suppose to be %t, got %t", i, testCase.transaction, parsed.UseTransaction)
		}
		if !reflect.DeepEqual(parsed.Statements, testCase.result) {
			t.Errorf("Parsed result for test case %d did not match expected results", i)
		}
	}
}

func TestNoTransactionMustBeInFirstLine(t *testing.T) {
	testMigration := `
	CREATE TABLE test_table1 (id integer not null primary key);

	-- +migration NoTransaction
	`

	reader := bytes.NewReader([]byte(testMigration))
	_, err := Parse(reader)
	if err == nil {
		t.Error("Expected parser to return error if -- +migration noTransaction was not the first line, but got no error")
	}
}
