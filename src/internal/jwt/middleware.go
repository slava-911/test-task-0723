package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cristalhq/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/slava-911/test-task-0723/internal/config"
)

type key string

const ctxKey key = "user_id"

func Middleware(echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := strings.Split(c.Request().Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			c.Logger().Error("Malformed token")
			return echo.NewHTTPError(http.StatusUnauthorized, "the correct token is required for authorization")
		}

		c.Logger().Debug("create jwt verifier")
		jwtToken := authHeader[1]
		c.Echo().Logger.Info(fmt.Sprintf("jwtToken: %s", jwtToken))
		verifier, err := jwt.NewVerifierHS(jwt.HS256, []byte(config.GetConfig().JWT.Secret))
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		c.Logger().Debug("parse and verify token")
		newToken, err := jwt.Parse([]byte(jwtToken), verifier)
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		c.Logger().Debug("parse user claims")
		var uc UserClaims
		err = json.Unmarshal(newToken.Claims(), &uc)
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}
		if valid := uc.IsValidAt(time.Now()); !valid {
			c.Logger().Error("token has been expired")
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		ctx := context.WithValue(c.Request().Context(), ctxKey, uc.ID)
		c.SetRequest(c.Request().WithContext(ctx))
		//h(c)
		return nil
	}
}
