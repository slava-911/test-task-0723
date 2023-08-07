package service

import (
	"context"

	cmodel "github.com/slava-911/test-task-0723/internal/controller/http/model"
	"github.com/slava-911/test-task-0723/internal/domain/model"
)

type UserService interface {
	Create(ctx context.Context, req *cmodel.CreateUserDTO) (dmodel.User, error)
	GetOneByEmail(ctx context.Context, email, password string) (dmodel.User, error)
	GetOneById(ctx context.Context, id string) (dmodel.User, error)
	Update(ctx context.Context, id string, chFields map[string]string, oldPass string) error
	Delete(ctx context.Context, id string) error
}

type OrderService interface {
	Create(ctx context.Context, req *cmodel.CreateOrderDTO) (dmodel.Order, error)
	GetAllByUserId(ctx context.Context, id string, limit, offset int) (cmodel.OrdersResponse, error)
	GetOneById(ctx context.Context, id string) (dmodel.Order, error)
	Complete(ctx context.Context, id string) error
	AddProduct(ctx context.Context, productId, orderId string, quantity int) error
	DeleteProduct(ctx context.Context, productId, orderId string) error
}

type ProductService interface {
	Create(ctx context.Context, req *cmodel.CreateProductDTO) (dmodel.Product, error)
	GetOneById(ctx context.Context, id string) (dmodel.Product, error)
	Update(ctx context.Context, req *cmodel.UpdateProductDTO) error
}
