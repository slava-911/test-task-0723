package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slava-911/test-task-0723/internal/apperror"
	cmodel "github.com/slava-911/test-task-0723/internal/controller/http/model"
	"github.com/slava-911/test-task-0723/internal/domain/model"
	"github.com/slava-911/test-task-0723/internal/storage"
	"github.com/slava-911/test-task-0723/pkg/logging"
)

type orderService struct {
	storage storage.OrderStorage
	logger  *logging.Logger
}

func NewOrderService(s storage.OrderStorage, l *logging.Logger) *orderService {
	return &orderService{
		storage: s,
		logger:  l,
	}
}

func (s *orderService) Create(ctx context.Context, req *cmodel.CreateOrderDTO) (r *dmodel.Order, err error) {
	newOrder := req.ToOrder()
	newOrder.Id = uuid.New().String()
	newOrder.CreatedAt = time.Now()
	newOrder, err = s.storage.Create(ctx, newOrder)
	if err != nil {
		s.logger.Error(err)
		return r, err
	}
	return newOrder, nil
}

func (s *orderService) GetAllByUserId(ctx context.Context, id string, limit, offset int) (r *cmodel.OrdersResponse, err error) {
	orders, err := s.storage.FindAllByUserId(ctx, id, limit, offset)
	if err != nil {
		s.logger.Error(err)
		return r, err
	}
	r = &cmodel.OrdersResponse{
		Limit:  limit,
		Offset: offset,
		Orders: orders,
	}
	return r, nil
}

func (s *orderService) GetOneById(ctx context.Context, id string) (r *dmodel.Order, err error) {
	r, err = s.storage.FindOneById(ctx, id)
	if err != nil {
		s.logger.Error(err)
		return r, err
	}
	return r, nil
}

func (s *orderService) Complete(ctx context.Context, id string) error {
	if err := s.storage.Complete(ctx, id); err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to complete order, error: %w", err)
	}
	return nil
}

func (s *orderService) AddProduct(ctx context.Context, productId, orderId string, quantity int) error {
	if err := s.storage.AddProduct(ctx, productId, orderId, quantity); err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to add product (id %s) to order (id %s), error: %w", productId, orderId, err)
	}
	return nil
}

func (s *orderService) DeleteProduct(ctx context.Context, productId, orderId string) error {
	if err := s.storage.DeleteProduct(ctx, productId, orderId); err != nil {
		s.logger.Error(err)
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to delete product (id %s) from order (id %s), error: %w", productId, orderId, err)
	}
	return nil
}
