package storage

import (
	"context"

	"github.com/slava-911/test-task-0723/internal/domain/model"
)

type UserStorage interface {
	Create(ctx context.Context, req *dmodel.User) (*dmodel.User, error)
	FindOneByEmail(ctx context.Context, email string) (*dmodel.User, error)
	FindOneById(ctx context.Context, id string) (*dmodel.User, error)
	Update(ctx context.Context, id string, chFields map[string]string) error
	Delete(ctx context.Context, id string) error
}

type OrderStorage interface {
	Create(ctx context.Context, req *dmodel.Order) (*dmodel.Order, error)
	FindAllByUserId(ctx context.Context, id string, limit, offset int) ([]*dmodel.Order, error)
	FindOneById(ctx context.Context, id string) (*dmodel.Order, error)
	Complete(ctx context.Context, id string) error
	AddProduct(ctx context.Context, productId, orderId string, quantity int) error
	DeleteProduct(ctx context.Context, productId, orderId string) error
}

type ProductStorage interface {
	Create(ctx context.Context, req *dmodel.Product) (*dmodel.Product, error)
	FindOneById(ctx context.Context, id string) (*dmodel.Product, error)
	Update(ctx context.Context, req *dmodel.Product) error
}
