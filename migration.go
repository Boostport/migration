package migration

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"

	"github.com/Boostport/migration/parser"
)

// Direction type up/down
type Direction int

// String returns a string representation of the direction
func (d Direction) String() string {
	switch d {
	case Up:
		return "up"
	case Down:
		return "down"
	default:
		return "directionless"
	}
}

// Constants for direction
const (
	Up Direction = iota
	Down
)

var numberPrefixRegex = regexp.MustCompile(`^(\d+).*$`)

// Migration represents a migration, containing statements for migrating up and down.
type Migration struct {
	ID   string
	Up   *parser.ParsedMigration
	Down *parser.ParsedMigration
}

// PlannedMigration is a migration with a direction defined. This allows the driver to
// work out how to apply the migration.
type PlannedMigration struct {
	*Migration
	Direction Direction
}

// Less compares two migrations to determine how they should be ordered.
func (m Migration) Less(other *Migration) bool {
	switch {
	case m.isNumeric() && other.isNumeric() && m.VersionInt() != other.VersionInt():
		return m.VersionInt() < other.VersionInt()
	case m.isNumeric() && !other.isNumeric():
		return true
	case !m.isNumeric() && other.isNumeric():
		return false
	default:
		return m.ID < other.ID
	}
}

func (m Migration) isNumeric() bool {
	return len(m.NumberPrefixMatches()) > 0
}

// NumberPrefixMatches returns a list of string matches
func (m Migration) NumberPrefixMatches() []string {
	return numberPrefixRegex.FindStringSubmatch(m.ID)
}

// VersionInt converts the migration version to an 64-bit integer.
func (m Migration) VersionInt() int64 {
	v := m.NumberPrefixMatches()[1]
	value, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Could not parse %q into int64: %s", v, err))
	}
	return value
}

type byID []*Migration

func (b byID) Len() int           { return len(b) }
func (b byID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byID) Less(i, j int) bool { return b[i].Less(b[j]) }

// Migrate runs a migration using a given driver and MigrationSource. The direction defines whether
// the migration is up or down, and max is the maximum number of migrations to apply. If max is set to 0,
// then there is no limit on the number of migrations to apply.
func Migrate(driver Driver, migrations Source, direction Direction, max int) (int, error) {
	count := 0

	m, err := getMigrations(migrations)
	if err != nil {
		return count, err
	}

	appliedMigrations, err := driver.Versions()
	if err != nil {
		return count, err
	}

	migrationsToApply := planMigrations(m, appliedMigrations, direction, max)
	for _, plannedMigration := range migrationsToApply {
		logPrintf("Applying migration (%s) named '%s'...", direction.String(), plannedMigration.ID)

		err = driver.Migrate(plannedMigration)
		if err != nil {
			errorMessage := "Error while running migration " + plannedMigration.ID

			if plannedMigration.Direction == Up {
				errorMessage += " (up)"
			} else {
				errorMessage += " (down)"
			}
			return count, fmt.Errorf(errorMessage+": %s", err)
		}

		logPrintf("Applied migration (%s) named '%s'", direction.String(), plannedMigration.ID)
		count++
	}

	err = driver.Close()
	return count, err
}

func getMigrations(migrations Source) ([]*Migration, error) {
	var m []*Migration
	tempMigrations := map[string]*Migration{}

	files, err := migrations.ListMigrationFiles()
	if err != nil {
		return m, err
	}

	regex := regexp.MustCompile(`(\d*_.*)\.(up|down)\..*`)

	for _, file := range files {
		matches := regex.FindStringSubmatch(file)

		if len(matches) > 0 && file == matches[0] {
			id := matches[1]
			direction := matches[2]

			if _, ok := tempMigrations[id]; !ok {
				tempMigrations[id] = &Migration{
					ID: id,
				}
			}

			reader, err := migrations.GetMigrationFile(file)
			if err != nil {
				return m, fmt.Errorf("Error getting migrations: %s", err)
			}

			contents, err := ioutil.ReadAll(reader)
			if err != nil {
				return m, fmt.Errorf("Error getting migration content: %s", err)
			}

			parsed, err := parser.Parse(bytes.NewReader(contents))
			if err != nil {
				return m, fmt.Errorf("Error parsing migration %s: %s", id, err)
			}

			if direction == "up" {
				tempMigrations[id].Up = parsed
			} else {
				tempMigrations[id].Down = parsed
			}
		}
	}

	for _, migration := range tempMigrations {
		m = append(m, migration)
	}

	sort.Sort(byID(m))

	return m, nil
}

func planMigrations(migrations []*Migration, appliedMigrations []string, direction Direction, max int) []*PlannedMigration {
	var applied []*Migration

	for _, appliedMigration := range appliedMigrations {
		applied = append(applied, &Migration{
			ID: appliedMigration,
		})
	}

	sort.Sort(byID(applied))

	// Get last migration that was run
	record := &Migration{}

	if len(applied) > 0 {
		record = applied[len(applied)-1]
	}

	var result []*PlannedMigration

	// Add missing migrations up to the last run migration.
	// This can happen for example when merges happened.
	if len(applied) > 0 {
		result = append(result, toCatchup(migrations, applied, record)...)
	}

	// Figure out which migrations to apply
	toApply := toApply(migrations, record.ID, direction)
	toApplyCount := len(toApply)

	if max > 0 && max < toApplyCount {
		toApplyCount = max
	}

	for _, v := range toApply[0:toApplyCount] {
		result = append(result, &PlannedMigration{
			Migration: v,
			Direction: direction,
		})
	}

	return result
}

// Filter a slice of migrations into ones that should be applied.
func toApply(migrations []*Migration, current string, direction Direction) []*Migration {
	var index = -1

	if current != "" {
		for index < len(migrations)-1 {
			index++
			if migrations[index].ID == current {
				break
			}
		}
	}

	if direction == Up {
		return migrations[index+1:]
	} else if direction == Down {
		if index == -1 {
			return []*Migration{}
		}

		// Add in reverse order
		toApply := make([]*Migration, index+1)
		for i := 0; i < index+1; i++ {
			toApply[index-i] = migrations[i]
		}
		return toApply
	}

	panic("Not possible")
}

// Get migrations that we need to apply regardless of whether the direction is up or down. This is
// because there may be migration "holes" due to merges.
func toCatchup(migrations, existingMigrations []*Migration, lastRun *Migration) []*PlannedMigration {
	var missing []*PlannedMigration

	for _, migration := range migrations {
		found := false

		for _, existing := range existingMigrations {
			if existing.ID == migration.ID {
				found = true
				break
			}
		}

		if !found && migration.Less(lastRun) {
			missing = append(missing, &PlannedMigration{Migration: migration, Direction: Up})
		}
	}

	return missing
}
