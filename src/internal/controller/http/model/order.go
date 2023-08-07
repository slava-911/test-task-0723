package cmodel

import (
	dmodel "github.com/slava-911/test-task-0723/internal/domain/model"
)

type CreateOrderDTO struct {
	UserId string `json:"user_id"`
}

func (d *CreateOrderDTO) ToOrder() *dmodel.Order {
	return &dmodel.Order{
		UserId: d.UserId,
	}
}

type OrdersResponse struct {
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	Orders []dmodel.Order `json:"orders"`
}
