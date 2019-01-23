// Copyright 2018 The SQLer Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"strings"

	"github.com/labstack/echo"
)

// routeIndex - the index route
func routeIndex(c echo.Context) error {
	return c.JSON(200, map[string]interface{}{
		"success": true,
		"message": "Welcome!",
	})
}

// routeExecMacro - execute the requested macro
func routeExecMacro(c echo.Context) error {
	macro := c.Get("macro").(*Macro)
	input := make(map[string]interface{})
	body := make(map[string]interface{})

	c.Bind(&body)

	for k := range c.QueryParams() {
		input[k] = c.QueryParam(k)
	}

	for k, v := range body {
		input[k] = v
	}

	headers := c.Request().Header
	for k, v := range headers {
		input["http_"+strings.Replace(strings.ToLower(k), "-", "_", -1)] = v[0]
	}

	out, err := macro.Call(input)
	if err != nil {
		code := errStatusCodeMap[err]
		if code < 1 {
			code = 500
		}
		return c.JSON(code, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"data":    out,
		})
	}

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"data":    out,
	})
}
