package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/slava-911/test-task-0723/internal/apperror"
	cmodel "github.com/slava-911/test-task-0723/internal/controller/http/model"
	"github.com/slava-911/test-task-0723/internal/domain/service"
	"github.com/slava-911/test-task-0723/internal/jwt"
	"github.com/slava-911/test-task-0723/pkg/logging"
)

const (
	ordersPath         = "/orders"
	ordersIdPath       = "/orders/:order_id"
	ordersCompletePath = "/orders/complete/:order_id"
	ordersContentPath  = "/orders/content/:order_id"
)

type orderHandler struct {
	orderService service.OrderService
	logger       *logging.Logger
}

func NewOrderHandler(s service.OrderService, l *logging.Logger) *orderHandler {
	return &orderHandler{
		orderService: s,
		logger:       l,
	}
}

func (h *orderHandler) Register(e *echo.Echo) {
	e.POST(ordersPath, h.CreateOrder, jwt.Middleware)
	e.GET(ordersPath, h.GetOrders, jwt.Middleware)
	e.GET(ordersIdPath, h.GetOrder, jwt.Middleware)
	e.POST(ordersCompletePath, h.CompleteOrder, jwt.Middleware)
	e.POST(ordersContentPath, h.AddProductToOrder, jwt.Middleware)
	e.DELETE(ordersContentPath, h.DeleteProductFromOrder, jwt.Middleware)
}

func (h *orderHandler) CreateOrder(c echo.Context) error {
	h.logger.Info("request received to create order")

	var req cmodel.CreateOrderDTO
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to decode order data: %w", err).Error())
	}

	resp, err := h.orderService.Create(c.Request().Context(), &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to create order: %w", err).Error())
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *orderHandler) GetOrders(c echo.Context) error {
	h.logger.Info("request received to get orders")

	pId := c.Request().Context().Value("user_id")
	if pId == nil {
		h.logger.Error("there is no user_id in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to parse parameter user_id")
	}
	userId := pId.(string)

	var (
		limit, offset int
		err           error
	)
	pLimit := c.QueryParam("limit")
	pOffset := c.QueryParam("offset")
	if pLimit == "" || pOffset == "" {
		limit = 1
		offset = 0
	} else {
		limit, err = strconv.Atoi(pLimit)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse parameter limit: %w", err).Error())
		}
		offset, err = strconv.Atoi(pOffset)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse parameter offset: %w", err).Error())
		}
	}

	resp, err := h.orderService.GetAllByUserId(c.Request().Context(), userId, limit, offset)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to get orders: %w", err)
		switch {
		case errors.Is(err, apperror.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, wrappedErr.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, wrappedErr.Error())
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *orderHandler) GetOrder(c echo.Context) error {
	h.logger.Info("request received to get order")

	orderId := c.Param("order_id")
	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse parameter order_id")
	}
	resp, err := h.orderService.GetOneById(c.Request().Context(), orderId)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to get order with id %s: %w", orderId, err)
		switch {
		case errors.Is(err, apperror.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, wrappedErr.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, wrappedErr.Error())
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *orderHandler) CompleteOrder(c echo.Context) error {
	h.logger.Info("request received to complete order")

	orderId := c.Param("order_id")
	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse parameter order_id")
	}
	err := h.orderService.Complete(c.Request().Context(), orderId)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to complete order with id %s: %w", orderId, err)
		switch {
		case errors.Is(err, apperror.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, wrappedErr.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, wrappedErr.Error())
		}
	}

	return c.JSON(http.StatusOK, "order completed")
}

func (h *orderHandler) AddProductToOrder(c echo.Context) error {
	h.logger.Info("request received to assign orders")

	orderId := c.Param("order_id")
	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse parameter order_id")
	}
	productId := c.QueryParam("product_id")
	if productId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse parameter product_id")
	}
	pQuantity := c.QueryParam("quantity")
	if pQuantity == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse parameter quantity")
	}
	quantity, err := strconv.Atoi(pQuantity)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse parameter quantity: %w", err).Error())
	}

	err = h.orderService.AddProduct(c.Request().Context(), productId, orderId, quantity)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to add product (id %s) to order (id %s): %w", productId, orderId, err)
		switch {
		case errors.Is(err, apperror.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, wrappedErr.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, wrappedErr.Error())
		}
	}

	return c.JSON(http.StatusOK, "product added")
}

func (h *orderHandler) DeleteProductFromOrder(c echo.Context) error {
	h.logger.Info("request received to assign orders")

	orderId := c.Param("order_id")
	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse parameter order_id")
	}
	productId := c.QueryParam("product_id")
	if productId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse parameter product_id")
	}

	err := h.orderService.DeleteProduct(c.Request().Context(), productId, orderId)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to delete product (id %s) from order (id %s): %w", productId, orderId, err)
		switch {
		case errors.Is(err, apperror.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, wrappedErr.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, wrappedErr.Error())
		}
	}

	return c.JSON(http.StatusOK, "product deleted")
}
