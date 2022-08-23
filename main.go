package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// go:embed public
var f embed.FS

func main() {
	router := gin.Default()

	router.StaticFile("/", "./public/index.html")

	router.Static("/public", "./public")

	router.StaticFS("/fs", http.FileSystem(http.FS(f)))

	router.GET("/employees", func(ctx *gin.Context) {
		ctx.File("./public/employee.html")
	})

	router.POST("/employees", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "New POST request success")
	})

	log.Fatal(router.Run(":3000"))
}
