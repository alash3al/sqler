// Copyright 2018 The SQLer Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"fmt"

	"github.com/alash3al/go-color"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()
	e.HideBanner = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORS())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 9}))
	e.Use(middleware.Recover())

	e.GET("/", routeIndex)
	e.Any("/:macro", routeExecMacro, middlewareAuthorize)

	fmt.Println(color.MagentaString(sqlerBrand))
	fmt.Printf("⇨ used dsn is %s \n", color.GreenString(*flagDBDSN))

	color.Red(e.Start(*flagListenAddr).Error())
}
