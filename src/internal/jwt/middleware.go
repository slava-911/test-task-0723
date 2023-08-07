package jwt

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cristalhq/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/slava-911/test-task-0723/internal/config"
	"github.com/slava-911/test-task-0723/pkg/logging"
)

type key string

const ctxKey key = "user_id"

func Middleware(h echo.HandlerFunc, logger *logging.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := strings.Split(c.Request().Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			logger.Error("Malformed token")
			return echo.NewHTTPError(http.StatusUnauthorized, "the correct token is required for authorization")
		}

		logger.Debug("create jwt verifier")
		jwtToken := authHeader[1]
		verifier, err := jwt.NewVerifierHS(jwt.HS256, []byte(config.GetConfig().JWT.Secret))
		if err != nil {
			logger.Error(err)
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		logger.Debug("parse and verify token")
		newToken, err := jwt.Parse([]byte(jwtToken), verifier)
		if err != nil {
			logger.Error(err)
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		logger.Debug("parse user claims")
		var uc UserClaims
		err = json.Unmarshal(newToken.Claims(), &uc)
		if err != nil {
			logger.Error(err)
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}
		if valid := uc.IsValidAt(time.Now()); !valid {
			logger.Error("token has been expired")
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		//c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), ctxKey, uc.ID)))
		c.Set("user_id", uc.ID)
		return h(c)
	}
}
