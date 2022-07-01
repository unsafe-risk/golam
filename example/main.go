package main

import (
	"github.com/unsafe-risk/golam"
	"net/http"
)

type Json map[string]interface{}

func main() {
	g := go....lam.New()
	g.GET("/", func(c golam.Context) error {
		return c.JSON(http.StatusOK, Json{
			"hello": "world",
		})
	})
	g.Any("/", func(c golam.Context) error {
		return c.JSON(http.StatusOK, Json{
			"hello": "world any",
		})
	})
	g.POST("/", func(c golam.Context) error {
		return c.JSON(http.StatusOK, Json{
			"hello": "world post",
		})
	})
	g.GET("/tt", func(c golam.Context) error {
		return c.JSON(http.StatusOK, Json{
			"hello tt": "world get",
		})
	})
	g.GET("/{param+}", func(c golam.Context) error {
		return c.JSON(http.StatusOK, Json{
			"hello": c.PathParams()["param"],
		})
	})
	g.StartWithLocalAddr(":3000")
}
