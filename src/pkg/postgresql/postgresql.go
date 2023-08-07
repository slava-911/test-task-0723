package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/slava-911/test-task-0723/pkg/logging"
	"github.com/slava-911/test-task-0723/pkg/utils"
)

type Client interface {
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Close()
}

// NewConnection establishes a database connection
func NewClient(ctx context.Context, connString string, maxAttempts int, connTimeout time.Duration) (pool *pgxpool.Pool, err error) {
	logger := logging.LoggerFromContext(ctx)

	var cfg *pgxpool.Config
	cfg, err = pgxpool.ParseConfig(connString)
	if err != nil {
		logger.Errorf("Unable to parse config: %v\n", err)
		return pool, err
	}

	err = utils.DoWithAttempts(func() error {
		ctx, cancel := context.WithTimeout(ctx, connTimeout)
		defer cancel()

		pool, err = pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			logger.Error("Failed to create pgx pool. Going to do the next attempt")
			return err
		}

		err = pool.Ping(ctx)
		if err != nil {
			logger.Error("Failed to ping database. Going to do the next attempt")
			return err
		}

		return nil
	}, maxAttempts, connTimeout)

	if err != nil {
		logger.Error("All attempts are exceeded. Unable to connect to postgres")
		return pool, err
	}

	return pool, nil
}

func DetailedPgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		pgErr = err.(*pgconn.PgError)
		newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
			pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		return newErr
	}
	return nil
}
