package middleware

import (
	"fmt"
	"time"

	"github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/marshaler"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

type CachedPage struct {
	URL        string
	HTML       []byte
	StatusCode int
	Headers    map[string]string
}

func PageCache(ch *cache.Cache) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.Request().URL.String()
			res, err := marshaler.New(ch).Get(c.Request().Context(), key, new(CachedPage))
			if err != nil {
				if err == redis.Nil {
					c.Logger().Infof("no cached page for: %s", key)
				} else {
					c.Logger().Errorf("failed getting cached page: %s", key)
					c.Logger().Error(err)
				}
				return next(c)
			}

			page := res.(*CachedPage)

			if page.Headers != nil {
				for k, v := range page.Headers {
					c.Response().Header().Set(k, v)
				}
			}
			c.Logger().Infof("serving cached page for: %s", key)

			return c.HTMLBlob(page.StatusCode, page.HTML)
		}
	}
}

func CacheControl(maxAge time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			v := "no-cache, no-store"
			if maxAge > 0 {
				v = fmt.Sprintf("public, max-age=%.0f", maxAge.Seconds())
			}
			c.Response().Header().Set("Cache-Control", v)
			return next(c)
		}
	}
}
