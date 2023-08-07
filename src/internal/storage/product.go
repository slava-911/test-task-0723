package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/slava-911/test-task-0723/internal/apperror"
	dmodel "github.com/slava-911/test-task-0723/internal/domain/model"
	"github.com/slava-911/test-task-0723/pkg/logging"
	"github.com/slava-911/test-task-0723/pkg/postgresql"
)

type productStorage struct {
	db     postgresql.Client
	logger *logging.Logger
}

func NewProductStorage(c postgresql.Client, l *logging.Logger) *productStorage {
	return &productStorage{
		db:     c,
		logger: l,
	}
}

func (s *productStorage) Create(ctx context.Context, p *dmodel.Product) (r dmodel.Product, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		INSERT INTO products
			(id, price, quantity, description, tags)
		VALUES
			($1, $2, $3, $4, $5)
		RETURNING id`

	s.logger.Trace("executing SQL query to create product")

	row := s.db.QueryRow(ctx, q, p.Id, p.Price, p.Quantity, p.Description, p.Tags)
	if err = row.Scan(&r.Id); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return r, detErr
		}
		return r, err
	}

	return *p, nil
}

func (s *productStorage) FindOneById(ctx context.Context, id string) (r dmodel.Product, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		SELECT
		    p.id, p.price, p.quantity, p.description, p.tags
		FROM
		    products p
		WHERE
		    p.id = $1`

	s.logger.Trace("executing SQL query to find product by id")

	row := s.db.QueryRow(ctx, q, id)
	if err = row.Scan(&r.Id, &r.Price, &r.Quantity, &r.Description, &r.Tags); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r, apperror.ErrNotFound
		}
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return r, detErr
		}
		return r, err
	}
	return r, nil
}

func (s *productStorage) Update(ctx context.Context, p *dmodel.Product) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		UPDATE
		    products
		SET
		    price = $2, quantity = $3, description = $3, tags = $4
		WHERE
		    id = $1`

	s.logger.Trace("executing SQL query to update product")

	if _, err := s.db.Exec(ctx, q, p.Id, p.Price, p.Quantity, p.Description, p.Tags); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}

	return nil
}
