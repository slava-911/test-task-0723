package metric

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	homePath      = "/"
	heartbeatPath = "/heartbeat"
)

// Register adds the routes for the metrics to the passed router.
func Register(e *echo.Echo, serviceName string) {
	e.GET(homePath, func(c echo.Context) error {
		//c.String(http.StatusOK, serviceName)
		return c.HTML(http.StatusOK, fmt.Sprintf("Hello! It is %s service", serviceName))
	})

	e.GET(heartbeatPath, func(c echo.Context) error {
		//return c.JSON(http.StatusOK, struct{ Status string }{Status: "OK"})
		return c.String(http.StatusOK, "OK")
	})
}
