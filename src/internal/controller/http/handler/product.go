package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/slava-911/test-task-0723/internal/apperror"
	cmodel "github.com/slava-911/test-task-0723/internal/controller/http/model"
	"github.com/slava-911/test-task-0723/internal/domain/service"
	"github.com/slava-911/test-task-0723/pkg/logging"
)

const (
	productsPath   = "/products"
	productsIdPath = "/products/:product_id"
)

type productHandler struct {
	productService service.ProductService
	logger         *logging.Logger
}

func NewProductHandler(s service.ProductService, l *logging.Logger) *productHandler {
	return &productHandler{
		productService: s,
		logger:         l,
	}
}

func (h *productHandler) Register(e *echo.Echo) {
	e.POST(productsPath, h.CreateProduct)
	e.PUT(productsPath, h.UpdateProduct)
	e.GET(productsIdPath, h.GetProduct)
}

func (h *productHandler) CreateProduct(c echo.Context) error {
	h.logger.Info("request received to create product")

	var req cmodel.CreateProductDTO
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to decode product data: %w", err).Error())
	}

	resp, err := h.productService.Create(c.Request().Context(), &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to create product: %w", err).Error())
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *productHandler) GetProduct(c echo.Context) error {
	h.logger.Info("request received to get product")

	productId := c.Param("product_id")
	if productId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse parameter product_id")
	}
	resp, err := h.productService.GetOneById(c.Request().Context(), productId)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to get product with id %s: %w", productId, err)
		switch {
		case errors.Is(err, apperror.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, wrappedErr.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, wrappedErr.Error())
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *productHandler) UpdateProduct(c echo.Context) error {
	h.logger.Info("request received to update product")

	var req cmodel.UpdateProductDTO
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to decode product data: %w", err).Error())
	}

	err = h.productService.Update(c.Request().Context(), &req)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to get product with id %s: %w", req.Id, err)
		switch {
		case errors.Is(err, apperror.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, wrappedErr.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, wrappedErr.Error())
		}
	}

	return c.JSON(http.StatusOK, "product updated")
}
