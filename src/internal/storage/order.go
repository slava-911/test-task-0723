package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/slava-911/test-task-0723/internal/apperror"
	dmodel "github.com/slava-911/test-task-0723/internal/domain/model"
	"github.com/slava-911/test-task-0723/pkg/logging"
	"github.com/slava-911/test-task-0723/pkg/postgresql"
)

type orderStorage struct {
	db     postgresql.Client
	logger *logging.Logger
}

func NewOrderStorage(c postgresql.Client, l *logging.Logger) *orderStorage {
	return &orderStorage{
		db:     c,
		logger: l,
	}
}

func (s *orderStorage) Create(ctx context.Context, o *dmodel.Order) (r *dmodel.Order, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		INSERT INTO orders
			(id, user_id, created_at, completed)
		VALUES
			($1, $2, $3, $4)
		RETURNING id`

	s.logger.Trace("executing SQL query to create order")

	row := s.db.QueryRow(ctx, q, o.Id, o.UserId, o.CreatedAt, false)
	if err = row.Scan(&o.Id); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return r, detErr
		}
		return r, err
	}

	return o, nil
}

func (s *orderStorage) FindAllByUserId(ctx context.Context, id string, limit, offset int) (r []*dmodel.Order, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		SELECT
			uo.id, uo.user_id, uo.created_at, uo.completed, SUM(oc.price * oc.quantity) AS cost
		FROM 
			(SELECT *
			FROM orders
			WHERE orders.user_id = $1
			ORDER BY orders.id
			LIMIT $2 OFFSET $3) AS uo
		LEFT JOIN orders_content oc
		ON uo.id = oc.order_id
		GROUP BY 
			uo.id, uo.user_id, uo.created_at, uo.completed`

	s.logger.Trace("executing SQL query to find all orders by user id")

	rows, err := s.db.Query(ctx, q, id, limit, offset)
	if err != nil {
		return r, err
	}

	r = make([]*dmodel.Order, 0, limit)
	for rows.Next() {
		var o dmodel.Order
		err = rows.Scan(&o.Id, &o.UserId, &o.CreatedAt, &o.Completed, &o.Cost)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return r, apperror.ErrNotFound
			}
			if detErr := postgresql.DetailedPgError(err); detErr != nil {
				return r, detErr
			}
			return r, err
		}
		r = append(r, &o)
	}

	if len(r) == 0 {
		return r, apperror.ErrNotFound
	}
	return r, nil
}

func (s *orderStorage) FindOneById(ctx context.Context, id string) (r *dmodel.Order, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		SELECT
			o.id, o.user_id, o.created_at, o.completed, SUM(oc.price * oc.quantity) AS cost
		FROM 
			orders o
		LEFT JOIN orders_content oc
		ON o.id = oc.order_id
		WHERE 
		    o.id = $1
		GROUP BY 
			o.id, o.user_id, o.created_at, o.completed`

	s.logger.Trace("executing SQL query to find order by id")

	var o dmodel.Order
	row := s.db.QueryRow(ctx, q, id)
	err = row.Scan(&o.Id, &o.UserId, &o.CreatedAt, &o.Completed, &o.Cost)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r, apperror.ErrNotFound
		}
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return r, detErr
		}
		return r, err
	}

	if o.Id == "" {
		return r, apperror.ErrNotFound
	}
	return &o, nil
}

func (s *orderStorage) Complete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		UPDATE 
		    orders
		SET 
		    completed = true
		WHERE 
		    id = $1`

	s.logger.Trace("executing SQL query to complete order")

	if _, err := s.db.Exec(ctx, q, id); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}
	return nil
}

func (s *orderStorage) AddProduct(ctx context.Context, productId, orderId string, quantity int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}
	s.logger.Trace("Started Tx to add product to order")

	q := `
		SELECT
		    p.id, p.price, p.quantity
		FROM
		    products p
		WHERE
		    p.id = $1`

	var p dmodel.Product
	row := tx.QueryRow(ctx, q, productId)
	if err = row.Scan(&p.Id, &p.Price, &p.Quantity); err != nil {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.ErrNotFound
		}
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}
	if p.Quantity < quantity {
		tx.Rollback(ctx)
		return fmt.Errorf("insufficient quantity in stock")
	}

	q = `
		INSERT INTO orders_content
			(order_id, product_id, price, quantity)
		VALUES
			($1, $2, $3, $4)
		RETURNING product_id`

	row = tx.QueryRow(ctx, q, orderId, productId, p.Price, quantity)
	if err = row.Scan(&p.Id); err != nil {
		tx.Rollback(ctx)
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}

	q = `
		UPDATE 
		    products
		SET 
		    quantity = $1
		WHERE 
		    id = $2`

	if _, err = tx.Exec(ctx, q, p.Quantity-quantity, productId); err != nil {
		tx.Rollback(ctx)
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}
	return nil
}

func (s *orderStorage) DeleteProduct(ctx context.Context, productId, orderId string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		DELETE FROM
		    orders_content oc
		WHERE
		    oc.order_id = $1 AND oc.product_id = $2`

	s.logger.Trace("executing SQL query to delete product from order")

	if _, err := s.db.Exec(ctx, q, orderId, productId); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}
	return nil
}
