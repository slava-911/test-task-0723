package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/slava-911/test-task-0723/internal/apperror"
	cmodel "github.com/slava-911/test-task-0723/internal/controller/http/model"
	dmodel "github.com/slava-911/test-task-0723/internal/domain/model"
	"github.com/slava-911/test-task-0723/internal/storage"
	"github.com/slava-911/test-task-0723/pkg/logging"
)

type userService struct {
	storage storage.UserStorage
	logger  *logging.Logger
}

func NewUserService(s storage.UserStorage, l *logging.Logger) *userService {
	return &userService{
		storage: s,
		logger:  l,
	}
}

func (s *userService) Create(ctx context.Context, req *cmodel.CreateUserDTO) (r *dmodel.User, err error) {
	newUser := req.ToUser()
	newUser.Id = uuid.New().String()
	s.logger.Debug("generate password hash")
	if err = newUser.GeneratePasswordHash(); err != nil {
		s.logger.Errorf("failed to create user due to error %v", err)
		return r, err
	}

	newUser, err = s.storage.Create(ctx, newUser)
	if err != nil {
		s.logger.Error(err)
		return r, err
	}
	return newUser, nil
}

func (s *userService) GetOneByEmail(ctx context.Context, email, pwd string) (r *dmodel.User, err error) {
	r, err = s.storage.FindOneByEmail(ctx, email)
	if err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return r, err
		}
		return r, fmt.Errorf("failed to find user by email, error: %w", err)
	}

	if err = r.CheckPassword(pwd); err != nil {
		return r, apperror.ErrNotFound
	}
	return r, nil
}

func (s *userService) GetOneById(ctx context.Context, id string) (r *dmodel.User, err error) {
	r, err = s.storage.FindOneById(ctx, id)
	if err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return r, err
		}
		return r, fmt.Errorf("failed to find user by id, error: %w", err)
	}
	return r, nil
}

func (s *userService) Update(ctx context.Context, id string, chFields map[string]string, oldPass string) error {

	if oldPass != "" {
		s.logger.Debug("get user by uuid")
		user, err := s.GetOneById(ctx, id)
		if err != nil {
			return err
		}

		s.logger.Debug("compare hash current password and old password")
		if err = user.CheckPassword(oldPass); err != nil {
			return fmt.Errorf("old password does not match current password")
		}

		user.Password = chFields["password"]

		s.logger.Debug("generate password hash")
		if err = user.GeneratePasswordHash(); err != nil {
			return fmt.Errorf("failed to update user, error %w", err)
		}

		chFields["password"] = user.Password
	}

	if err := s.storage.Update(ctx, id, chFields); err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to update user, error: %w", err)
	}

	return nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	if err := s.storage.Delete(ctx, id); err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to delete user, error: %w", err)
	}
	return nil
}
