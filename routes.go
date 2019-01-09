// Copyright 2018 The SQLer Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"github.com/labstack/echo"
)

// routeIndex - the index route
func routeIndex(c echo.Context) error {
	return c.JSON(200, map[string]interface{}{
		"success": true,
		"message": "Weclome!",
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

	out, err := macro.Call(input)
	if err != nil {
		return c.JSON(500, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"data":    out,
	})
}
