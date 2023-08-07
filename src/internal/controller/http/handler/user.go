package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/slava-911/test-task-0723/internal/apperror"
	cmodel "github.com/slava-911/test-task-0723/internal/controller/http/model"
	dmodel "github.com/slava-911/test-task-0723/internal/domain/model"
	"github.com/slava-911/test-task-0723/internal/domain/service"
	"github.com/slava-911/test-task-0723/internal/jwt"
	"github.com/slava-911/test-task-0723/pkg/logging"
	"github.com/slava-911/test-task-0723/pkg/utils"
)

const (
	signUpPath = "/signup"
	authPath   = "/auth"
	userPath   = "/profile"
)

type userHandler struct {
	userService service.UserService
	jwtHelper   jwt.Helper
	validate    *validator.Validate
	logger      *logging.Logger
}

func NewUserHandler(s service.UserService, h jwt.Helper, v *validator.Validate, l *logging.Logger) *userHandler {
	return &userHandler{
		userService: s,
		jwtHelper:   h,
		validate:    v,
		logger:      l,
	}
}

func (h *userHandler) Register(e *echo.Echo) {
	e.POST(signUpPath, h.Signup)
	e.POST(authPath, h.Auth)
	e.PUT(authPath, h.Auth)
	e.GET(userPath, h.GetUser, jwt.Middleware)
	e.PATCH(userPath, h.PartiallyUpdateUser, jwt.Middleware)
	e.DELETE(userPath, h.DeleteUser, jwt.Middleware)
}

func (h *userHandler) Signup(c echo.Context) error {
	h.logger.Info("request to create user")

	var (
		err   error
		token []byte
		req   *cmodel.CreateUserDTO
		user  *dmodel.User
	)

	if err = c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to decode user data: %w", err).Error())
	}

	if err = h.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, utils.TranslateValidationError(err, ""))
	}
	if err = req.ValidatePassword(); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	if user, err = h.userService.Create(c.Request().Context(), req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to create user: %w", err).Error())
	}

	if token, err = h.jwtHelper.GenerateAccessToken(user); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, token)
}

func (h *userHandler) Auth(c echo.Context) error {
	h.logger.Info("authorization request")

	var (
		err   error
		token []byte
		req   *cmodel.SignInUserDTO
		user  *dmodel.User
		rt    jwt.RT
	)

	switch c.Request().Method {
	case http.MethodPost:
		if err = c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to decode user data: %w", err).Error())
		}

		if user, err = h.userService.GetOneByEmail(c.Request().Context(), req.Email, req.Password); err != nil {
			return err
		}

		if token, err = h.jwtHelper.GenerateAccessToken(user); err != nil {
			return err
		}
	case http.MethodPut:
		if err = c.Bind(&rt); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to decode token data: %w", err).Error())
		}

		if token, err = h.jwtHelper.UpdateRefreshToken(rt); err != nil {
			return err
		}
	}

	return c.JSON(http.StatusOK, token)
}

func (h *userHandler) GetUser(c echo.Context) error {
	h.logger.Info("request to get user")

	var (
		err  error
		user *dmodel.User
	)

	pId := c.Request().Context().Value("user_id")
	if pId == nil {
		h.logger.Error("there is no user_id in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to parse parameter user_id")
	}
	userId := pId.(string)

	if user, err = h.userService.GetOneById(c.Request().Context(), userId); err != nil {
		wrappedErr := fmt.Errorf("failed to get user with id %d: %w", userId, err)
		switch {
		case errors.Is(err, apperror.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, wrappedErr.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, wrappedErr.Error())
		}
	}

	return c.JSON(http.StatusOK, user)
}

func (h *userHandler) PartiallyUpdateUser(c echo.Context) error {
	h.logger.Info("request to update user")

	pId := c.Request().Context().Value("user_id")
	if pId == nil {
		h.logger.Error("there is no user_id in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to parse parameter user_id")
	}
	userId := pId.(string)

	var req cmodel.UpdateUserDTO
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to decode user data: %w", err).Error())
	}

	changedFields := make(map[string]string)
	oldPassword := ""
	if req.FirstName != nil {
		if err = h.validate.Var(*req.FirstName, "required,min=2"); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, utils.TranslateValidationError(err, "FirstName"))
		}
		changedFields["firstname"] = *req.FirstName
	}
	if req.LastName != nil {
		if err = h.validate.Var(*req.LastName, "required,min=2"); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, utils.TranslateValidationError(err, "LastName"))
		}
		changedFields["lastname"] = *req.LastName
	}
	if req.Email != nil {
		if err = h.validate.Var(*req.Email, "required,email"); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, utils.TranslateValidationError(err, "Email"))
		}
		changedFields["email"] = *req.Email
	}
	if req.OldPassword != nil && req.NewPassword != nil {
		if *req.OldPassword != *req.NewPassword && *req.OldPassword != "" && *req.NewPassword != "" {
			if err = h.validate.Var(*req.NewPassword, "required,min=8"); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, utils.TranslateValidationError(err, "NewPassword"))
			}
			oldPassword = *req.OldPassword
			changedFields["password"] = *req.NewPassword
		}
	}
	if len(changedFields) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Nothing to update")
	}

	err = h.userService.Update(c.Request().Context(), userId, changedFields, oldPassword)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to update user with id %s: %w", userId, err)
		switch {
		case errors.Is(err, apperror.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, wrappedErr.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, wrappedErr.Error())
		}
	}

	return c.JSON(http.StatusOK, "user updated")
}

func (h *userHandler) DeleteUser(c echo.Context) error {
	h.logger.Info("request to delete user")

	pId := c.Request().Context().Value("user_id")
	if pId == nil {
		h.logger.Error("there is no user_id in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to parse parameter user_id")
	}
	userId := pId.(string)

	err := h.userService.Delete(c.Request().Context(), userId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "user deleted")
}
