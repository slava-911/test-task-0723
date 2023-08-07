package cmodel

import dmodel "github.com/slava-911/test-task-0723/internal/domain/model"

type CreateProductDTO struct {
	Price       int      `json:"price"`
	Quantity    int      `json:"quantity"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func (d *CreateProductDTO) ToProduct() *dmodel.Product {
	return &dmodel.Product{
		Price:       d.Price,
		Quantity:    d.Quantity,
		Description: d.Description,
		Tags:        d.Tags,
	}
}

type UpdateProductDTO struct {
	Id          string   `json:"id"`
	Price       int      `json:"price"`
	Quantity    int      `json:"quantity"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func (d *UpdateProductDTO) ToProduct() *dmodel.Product {
	return &dmodel.Product{
		Id:          d.Id,
		Price:       d.Price,
		Quantity:    d.Quantity,
		Description: d.Description,
		Tags:        d.Tags,
	}
}

type ProductsResponse struct {
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	Products []*dmodel.Product `json:"products"`
}
