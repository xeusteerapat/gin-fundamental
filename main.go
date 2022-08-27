package main

import (
	"embed"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// go:embed public
var f embed.FS

type TimeoffRequest struct {
	Date   time.Time `json:"date" form:"date" binding:"required,future" time_format:"2006-01-02"`
	Amount float64   `json:"amount" form:"amount" binding:"required,gt=0"`
}

var ValidatorFutureDate validator.Func = func(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)

	if ok {
		return date.After(time.Now())
	}

	return true
}

func main() {
	router := gin.Default()

	// Binding validator
	if value, ok := binding.Validator.Engine().(*validator.Validate); ok {
		value.RegisterValidation("future", ValidatorFutureDate)
	}

	// Serving static files
	router.StaticFile("/", "./public/index.html")

	router.Static("/public", "./public")

	router.StaticFS("/fs", http.FileSystem(http.FS(f)))

	// Route params
	router.GET("/employees", func(ctx *gin.Context) {
		ctx.File("./public/employee.html")
	})

	// POST mothod, receive data from Form input
	router.POST("/employees", func(ctx *gin.Context) {
		var timeoffRequest TimeoffRequest

		if err := ctx.ShouldBind(&timeoffRequest); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		ctx.JSON(http.StatusOK, timeoffRequest)
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

	// Request query
	// eg. http://localhost:3000/query/?username=xeusteerapat&year=2022&month=8&month=9
	router.GET("/query/*rest", func(ctx *gin.Context) {
		username := ctx.Query("username")
		year := ctx.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
		months := ctx.QueryArray("month")

		ctx.IndentedJSON(http.StatusOK, gin.H{
			"username": username,
			"year":     year,
			"months":   months,
		})
	})

	apiGroup := router.Group("/api")
	apiGroup.POST("/timeoff", func(ctx *gin.Context) {
		var timeoffRequest TimeoffRequest

		if err := ctx.ShouldBind(&timeoffRequest); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		ctx.JSON(http.StatusOK, timeoffRequest)
	})

	log.Fatal(router.Run(":3000"))
}
