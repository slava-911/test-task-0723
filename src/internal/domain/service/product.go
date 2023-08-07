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

type productService struct {
	storage storage.ProductStorage
	logger  *logging.Logger
}

func NewProductService(s storage.ProductStorage, l *logging.Logger) *productService {
	return &productService{
		storage: s,
		logger:  l,
	}
}

func (s *productService) Create(ctx context.Context, req *cmodel.CreateProductDTO) (r *dmodel.Product, err error) {
	newProduct := req.ToProduct()
	newProduct.Id = uuid.New().String()
	newProduct, err = s.storage.Create(ctx, newProduct)
	if err != nil {
		s.logger.Error(err)
		return r, err
	}
	return newProduct, nil
}

func (s *productService) GetOneById(ctx context.Context, id string) (r *dmodel.Product, err error) {
	r, err = s.storage.FindOneById(ctx, id)
	if err != nil {
		s.logger.Error(err)
		return r, err
	}
	return r, nil
}

func (s *productService) Update(ctx context.Context, req *cmodel.UpdateProductDTO) error {
	product := req.ToProduct()
	if err := s.storage.Update(ctx, product); err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to update user, error: %w", err)
	}
	return nil
}
