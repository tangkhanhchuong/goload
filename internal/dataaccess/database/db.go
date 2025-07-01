package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"goload/internal/configs"
)

type Database interface {
	Delete(table interface{}) *goqu.DeleteDataset
	Dialect() string
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	From(from ...interface{}) *goqu.SelectDataset
	Insert(table interface{}) *goqu.InsertDataset
	Logger(logger goqu.Logger)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ScanStruct(i interface{}, query string, args ...interface{}) (bool, error)
	ScanStructContext(ctx context.Context, i interface{}, query string, args ...interface{}) (bool, error)
	ScanStructs(i interface{}, query string, args ...interface{}) error
	ScanStructsContext(ctx context.Context, i interface{}, query string, args ...interface{}) error
	ScanVal(i interface{}, query string, args ...interface{}) (bool, error)
	ScanValContext(ctx context.Context, i interface{}, query string, args ...interface{}) (bool, error)
	ScanVals(i interface{}, query string, args ...interface{}) error
	ScanValsContext(ctx context.Context, i interface{}, query string, args ...interface{}) error
	Select(cols ...interface{}) *goqu.SelectDataset
	Trace(op string, sqlString string, args ...interface{})
	Truncate(table ...interface{}) *goqu.TruncateDataset
	Update(table interface{}) *goqu.UpdateDataset
}

func InitializeDBAndMigrateUp(databaseConfig configs.Database, logger *zap.Logger) (*sql.DB, func(), error) {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		databaseConfig.Username,
		databaseConfig.Password,
		databaseConfig.Host,
		databaseConfig.Port,
		databaseConfig.Database,
	)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		logger.With(zap.Error(err)).Error("error connecting to the database")
		return nil, nil, err
	}

	cleanup := func() {
		db.Close()
	}

	migrator := NewMigrator(db, logger)
	err = migrator.Up(context.Background())
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to execute database up migration")
	}

	return db, cleanup, nil
}

func InitializeGoquDB(db *sql.DB) *goqu.Database {
	return goqu.New("postgres", db)
}
