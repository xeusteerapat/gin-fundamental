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

	// Serving static files
	router.StaticFile("/", "./public/index.html")

	router.Static("/public", "./public")

	router.StaticFS("/fs", http.FileSystem(http.FS(f)))

	// Route params
	router.GET("/employees", func(ctx *gin.Context) {
		ctx.File("./public/employee.html")
	})

	router.POST("/employees", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "New POST request success")
	})

	router.GET("/employees/:username/*rest", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"username": ctx.Param("username"),
			"rest":     ctx.Param("rest"),
		})
	})

	// Routing Groups
	// adminGroup := router.Group("/admin")

	// adminGroup.GET("/users", func(ctx *gin.Context) {
	// 	ctx.String(http.StatusOK, "Users Admin page")
	// })

	// adminGroup.GET("/users", func(ctx *gin.Context) {
	// 	ctx.String(http.StatusOK, "Roles Admin page")
	// })

	// adminGroup.GET("/users", func(ctx *gin.Context) {
	// 	ctx.String(http.StatusOK, "Policies Admin page")
	// })

	// Request objects
	router.GET("/request-object", func(ctx *gin.Context) {
		url := ctx.Request.URL.String()
		headers := ctx.Request.Header
		cookies := ctx.Request.Cookies()

		ctx.IndentedJSON(http.StatusOK, gin.H{
			"url":     url,
			"headers": headers,
			"cookies": cookies,
		})
	})

	log.Fatal(router.Run(":3000 🎉"))
}
