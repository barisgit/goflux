package echo

import (
	"embed"

	"github.com/barisgit/goflux"
	"github.com/labstack/echo/v4"
)

// StaticHandler creates an Echo handler using the shared static logic
func StaticHandler(assets embed.FS, config goflux.StaticConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		response := goflux.ServeStaticFile(assets, config, c.Request().URL.Path)

		if response.NotFound {
			return c.NoContent(404)
		}

		c.Response().Header().Set("Content-Type", response.ContentType)
		c.Response().Header().Set("Cache-Control", response.CacheControl)
		c.Response().WriteHeader(response.StatusCode)
		c.Response().Write(response.Body)
		return nil
	}
}
