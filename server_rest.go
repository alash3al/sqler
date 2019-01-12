// Copyright 2018 The SQLer Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// initialize RESTful server
func initRESTServer() error {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORS())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 9}))
	e.Use(middleware.Recover())

	e.GET("/", routeIndex)
	e.Any("/:macro", routeExecMacro, middlewareAuthorize)

	return e.Start(*flagRESTListenAddr)
}

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

		// for _, endpoint := range macro.Authorizers {
		// 	parts := strings.SplitN(endpoint, " ", 2)
		// 	if len(parts) < 2 {
		// 		return c.JSON(500, map[string]interface{}{
		// 			"success": false,
		// 			"error":   fmt.Sprintf("authorizer: %s is invalid", endpoint),
		// 		})
		// 	}
		// 	resp, err := resty.R().SetHeaders(map[string]string{
		// 		"Authorization": c.Request().Header.Get("Authorization"),
		// 	}).Execute(parts[0], parts[1])
		// 	if err != nil {
		// 		return c.JSON(500, map[string]interface{}{
		// 			"success": false,
		// 			"error":   err.Error(),
		// 		})
		// 	}
		// 	if resp.StatusCode() >= 400 {
		// 		return c.JSON(resp.StatusCode(), map[string]interface{}{
		// 			"success": false,
		// 			"error":   resp.Status(),
		// 		})
		// 	}
		// }

		c.Set("macro", macro)

		return next(c)
	}
}
