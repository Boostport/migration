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

// ParsedMigration is a parsed migration
type ParsedMigration struct {
	UseTransaction bool
	Statements     []string
}

func splitStatementsBySemicolon(buf string) []string {
	statements := strings.SplitAfter(buf, ";")

	for i, statement := range statements {
		trimmed := strings.TrimSpace(statement)
		if trimmed == "" && i > 0 {
			statements[i-1] += statements[i]
			statements = statements[:i+copy(statements[i:], statements[i+1:])]
		}
	}

	return statements
}

// Parse reads a migration and returns a parsed migrations
func Parse(r io.Reader) (*ParsedMigration, error) {
	p := &ParsedMigration{
		UseTransaction: true,
		Statements:     []string{},
	}

	var buf bytes.Buffer

	scanner := bufio.NewScanner(r)
	scanner.Split(scanLines)

	isFirstLine := true

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, sqlCmdPrefix) {
			option := strings.Replace(trimmed, sqlCmdPrefix, "", -1)

			switch option {
			case optionNoTransaction:
				if !isFirstLine && buf.Len() > 0 {
					return p, fmt.Errorf("%s%s must be in the first line of the migration", sqlCmdPrefix, optionNoTransaction)
				}
				p.UseTransaction = false

			case optionBeginStatement:
				// Add lines encountered before beginning the statement
				withoutCR := string(dropCR(buf.Bytes()))

				if !p.UseTransaction {
					p.Statements = append(p.Statements, splitStatementsBySemicolon(withoutCR)...)
				} else {
					p.Statements = append(p.Statements, withoutCR)
				}
				buf.Reset()

			case optionEndStatement:
				// Add the lines encountered during a statement block as 1 block
				p.Statements = append(p.Statements, string(dropCR(buf.Bytes())))

				buf.Reset()
			}
		} else if _, err := buf.WriteString(line); err != nil {
			return p, errors.New("error writing line to buffer")
		}

		isFirstLine = false
	}

	// If the buffer contains lines, process them
	if buf.Len() > 0 && strings.TrimSpace(buf.String()) != "" {
		withoutCR := string(dropCR(buf.Bytes()))

		if !p.UseTransaction {
			p.Statements = append(p.Statements, splitStatementsBySemicolon(withoutCR)...)
		} else {
			p.Statements = append(p.Statements, withoutCR)
		}
	}

	return p, nil
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
