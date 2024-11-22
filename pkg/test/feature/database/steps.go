package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/bool64/dbdog"
	"github.com/bool64/sqluct"
	"github.com/cucumber/godog"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

var errUnknownDatabase = errors.New("unknown database")

// ManagerCleaner clean database connections.
type ManagerCleaner struct {
	*dbdog.Manager
}

// RegisterSteps adds database manager context scenario steps to test suite.
func (m *ManagerCleaner) RegisterSteps(s *godog.ScenarioContext) {
	s.Step(`there is a clean "([^"]*)" database$`,
		m.thereIsACleanDatabase)
}

func (m *ManagerCleaner) thereIsACleanDatabase(dbName string) error {
	instance, ok := m.Instances[dbName]
	if !ok {
		return fmt.Errorf("%w: %s", errUnknownDatabase, dbName)
	}

	nonCleanTable := make([]string, 0)
	for tableName := range instance.Tables {
		nonCleanTable = append(nonCleanTable, tableName)
	}

	for len(nonCleanTable) > 0 {
		tableName := nonCleanTable[0]
		nonCleanTable = nonCleanTable[1:]

		var (
			postCleanup           bool
			postCleanupTableNames []string
		)

		if instance.PostCleanup != nil {
			postCleanup = true
			postCleanupTableNames = instance.PostCleanup[tableName]
		}

		err := m.cleanUpTable(tableName, dbName, instance.Storage, postCleanup, postCleanupTableNames)
		if err != nil {
			if isForeignKeyViolation(err) {
				nonCleanTable = append(nonCleanTable, tableName)

				continue
			}

			return err
		}
	}

	return nil
}

// isForeignKeyViolation checks whether the error is foreign key violation.
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == pgerrcode.ForeignKeyViolation
}

func (m *ManagerCleaner) cleanUpTable(tableName, dbName string, storage *sqluct.Storage, postCleanup bool, postCleanupTableNames []string) error {
	// Deleting from table
	_, err := storage.Exec(
		context.Background(),
		storage.DeleteStmt(tableName),
	)
	if err != nil {
		return fmt.Errorf("failed to delete from table %s in db %s: %w", tableName, dbName, err)
	}

	if postCleanup {
		for _, statement := range postCleanupTableNames {
			_, err = storage.DB().ExecContext(
				context.Background(),
				statement,
			)
			if err != nil {
				return fmt.Errorf(
					"failed to execute post cleanup statement %q for table %s in db %s: %w",
					statement, tableName, dbName, err,
				)
			}
		}
	}

	return nil
}
