package main

import (
	"strings"

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
			"error":   "resource not found #1",
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
		return c.JSON(404, map[string]interface{}{
			"success": false,
			"error":   "resource not found #2",
		})
	}

	input := make(map[string]interface{})
	body := make(map[string]interface{})

	for k := range c.QueryParams() {
		input[k] = c.QueryParam(k)
	}

	c.Bind(&body)

	for k, v := range body {
		input[k] = v
	}

	if len(macro.Rules) > 0 {
		result := Validate(input, macro.Rules)
		if len(result) > 0 {
			return c.JSON(422, map[string]interface{}{
				"success": false,
				"errors":  result,
			})
		}
	}

	out, err := macrosManager.Call(c.Param("macro"), input)
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
