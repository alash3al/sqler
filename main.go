package main

import (
	"log"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()
	e.HideBanner = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORS())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 9}))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", routeIndex)
	e.Any("/:macro", routeExecMacro)

	log.Fatal(e.Start(*flagListenAddr))
}
