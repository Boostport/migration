package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	sqlCmdPrefix         = "-- +migration "
	optionNoTransaction  = "NoTransaction"
	optionBeginStatement = "BeginStatement"
	optionEndStatement   = "EndStatement"
)

type ParsedMigration struct {
	UseTransaction bool
	Statements     []string
}

func splitStatementsBySemicolon(buf bytes.Buffer) []string {

	statements := strings.SplitAfter(buf.String(), ";")

	for i, statement := range statements {
		trimmed := strings.TrimSpace(statement)

		if trimmed == "" {
			statements = statements[:i+copy(statements[i:], statements[i+1:])]
		} else {
			statements[i] = trimmed + "\n"
		}
	}

	return statements
}

func Parse(r io.Reader) (*ParsedMigration, error) {

	p := &ParsedMigration{
		UseTransaction: true,
		Statements:     []string{},
	}

	var buf bytes.Buffer
	scanner := bufio.NewScanner(r)

	isFirstLine := true

	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, sqlCmdPrefix) {
			option := strings.Replace(line, sqlCmdPrefix, "", -1)

			switch option {
			case optionNoTransaction:
				if !isFirstLine && buf.Len() > 0 {
					return p, fmt.Errorf("%s%s must be in the first line of the migration.", sqlCmdPrefix, optionNoTransaction)
				}

				p.UseTransaction = false

			case optionBeginStatement:

				// Add lines encountered before beginning the statement
				if !p.UseTransaction {
					p.Statements = append(p.Statements, splitStatementsBySemicolon(buf)...)
				} else {
					p.Statements = append(p.Statements, buf.String())
				}

				buf.Reset()

			case optionEndStatement:

				// Add the lines encountered during a statement block as 1 block
				p.Statements = append(p.Statements, buf.String())

				buf.Reset()
			}
		} else if line != "" {
			if _, err := buf.WriteString(line + "\n"); err != nil {
				return p, errors.New("Error writing line to buffer")
			}
		}

		isFirstLine = false
	}

	// If the buffer contains lines, process them
	if buf.Len() > 0 {
		if !p.UseTransaction {
			p.Statements = append(p.Statements, splitStatementsBySemicolon(buf)...)
		} else {
			p.Statements = append(p.Statements, buf.String())
		}
	}

	return p, nil
}
