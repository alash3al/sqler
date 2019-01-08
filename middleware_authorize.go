package main

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty"
	"github.com/labstack/echo"
)

// middlewareAuthorize - the authorizer middleware
func middlewareAuthorize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if strings.HasPrefix(c.Param("macro"), "_") {
			return c.JSON(403, map[string]interface{}{
				"success": false,
				"error":   "access not allowed",
			})
		}

		macro := macrosManager.Get(c.Param("macro"))
		if macro == nil {
			return c.JSON(404, map[string]interface{}{
				"success": false,
				"error":   "resource not found",
			})
		}

		if len(macro.Methods) < 1 {
			macro.Methods = []string{c.Request().Method}
		}

		methodIsAllowed := false
		for _, method := range macro.Methods {
			method = strings.ToUpper(method)
			if c.Request().Method == method {
				methodIsAllowed = true
				break
			}
		}

		if !methodIsAllowed {
			return c.JSON(405, map[string]interface{}{
				"success": false,
				"error":   "method not allowed",
			})
		}

		for _, endpoint := range macro.Authorizers {
			parts := strings.SplitN(endpoint, " ", 2)
			if len(parts) < 2 {
				return c.JSON(500, map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("authorizer: %s is invalid", endpoint),
				})
			}
			resp, err := resty.R().SetHeaders(map[string]string{}).Execute(parts[0], parts[1])
			if err != nil {
				return c.JSON(500, map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				})
			}
			if resp.StatusCode() >= 400 {
				return c.JSON(resp.StatusCode(), map[string]interface{}{
					"success": false,
					"error":   resp.Status(),
				})
			}
		}

		c.Set("macro", macro)

		return next(c)
	}
}
