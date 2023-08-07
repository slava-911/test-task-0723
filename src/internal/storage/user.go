package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/slava-911/test-task-0723/internal/apperror"
	dmodel "github.com/slava-911/test-task-0723/internal/domain/model"
	"github.com/slava-911/test-task-0723/pkg/logging"
	"github.com/slava-911/test-task-0723/pkg/postgresql"
)

type userStorage struct {
	db     postgresql.Client
	logger *logging.Logger
}

func NewUserStorage(c postgresql.Client, l *logging.Logger) *userStorage {
	return &userStorage{
		db:     c,
		logger: l,
	}
}

func (s *userStorage) Create(ctx context.Context, u *dmodel.User) (r *dmodel.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		INSERT INTO users
			(id, firstname, lastname, email, password, age, is_married)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	s.logger.Trace("executing SQL query to create user")

	row := s.db.QueryRow(ctx, q, u.Id, u.FirstName, u.LastName, u.Email, u.Password, u.Age, u.IsMarried)
	if err = row.Scan(&u.Id); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return r, detErr
		}
		return r, err
	}

	return u, nil
}

func (s *userStorage) FindOneByEmail(ctx context.Context, email string) (r *dmodel.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		SELECT
		    u.id, u.firstname, u.lastname, u.email, u.password, u.age, u.is_married
		FROM
		    users u
		WHERE
		    u.email = $1`

	s.logger.Trace("executing SQL query to find user by email")

	row := s.db.QueryRow(ctx, q, email)
	if err = row.Scan(&r.Id, &r.FirstName, &r.LastName, &r.Email, &r.Password, &r.Age, &r.IsMarried); err != nil {
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

func (s *userStorage) FindOneById(ctx context.Context, id string) (r *dmodel.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		SELECT
		    u.id, u.firstname, u.lastname, u.email, u.password, u.age, u.is_married
		FROM
		    users u
		WHERE
		    u.id = $1`

	s.logger.Trace("executing SQL query to find user by id")

	row := s.db.QueryRow(ctx, q, id)
	if err = row.Scan(&r.Id, &r.FirstName, &r.LastName, &r.Email, &r.Password, &r.Age, &r.IsMarried); err != nil {
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

func (s *userStorage) Update(ctx context.Context, id string, chFields map[string]string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fields := make([]string, 0)
	params := make([]interface{}, 0)
	paramNum := 1

	for k, v := range chFields {
		fields = append(fields, fmt.Sprintf("%s=$%d", k, paramNum))
		params = append(params, v)
		paramNum++
	}

	fieldsToSet := strings.Join(fields, ", ")

	q := `
		UPDATE
		    users u
		SET
		    %s
		WHERE
		    u.id = $%d`

	q = fmt.Sprintf(q, fieldsToSet, paramNum)

	params = append(params, id)
	s.logger.Trace("executing SQL query to update user")
	s.logger.Tracef("params: %s", params)

	if _, err := s.db.Exec(ctx, q, params...); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}

	return nil
}

func (s *userStorage) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		DELETE FROM
		    users u
		WHERE
		    u.id = $1`

	s.logger.Trace("executing SQL query to delete user")

	if _, err := s.db.Exec(ctx, q, id); err != nil {
		if detErr := postgresql.DetailedPgError(err); detErr != nil {
			return detErr
		}
		return err
	}

	return nil
}
