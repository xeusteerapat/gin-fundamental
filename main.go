package main

import (
	"embed"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/xeusteerapat/gin-fundamental/employee"
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
	router.Use(ErrorMiddleware)

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

	router.StaticFile("/download", "./public/download.html")
	router.GET("/arsenal", func(ctx *gin.Context) {
		// ctx.File("./arsenal.txt") // render arsenal text content

		f, err := os.Open("./arsenal.txt")
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}

		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}

		ctx.Data(http.StatusOK, "text/plain", data)
	})

	// Get file stats and download the file
	router.GET("/teerapat", func(ctx *gin.Context) {
		// ctx.File("./teerapat.txt") // render teerapat.txt

		f, err := os.Open("./teerapat.txt")
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}

		defer f.Close()

		fStats, err := f.Stat()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}

		ctx.DataFromReader(http.StatusOK, fStats.Size(), "text/plain", f, map[string]string{
			"Content-Disposition": "attachment;filename=teerapat.txt",
		})
	})

	// Streaning data
	router.GET("/stream", func(ctx *gin.Context) {
		f, err := os.Open("./arsenal.txt")
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}

		defer f.Close()
		ctx.Stream(streamer(f))
	})

	// Using template
	router.LoadHTMLGlob("./templates/*")
	registerTemplateRoute(router)

	log.Fatal(router.Run(":3000"))
}

func streamer(r io.Reader) func(io.Writer) bool {
	return func(step io.Writer) bool {
		for {
			buf := make([]byte, 4*2^10)
			if _, err := r.Read(buf); err == nil {
				_, err := step.Write(buf)

				return err == nil
			} else {
				return false
			}
		}
	}
}

func registerTemplateRoute(r *gin.Engine) {
	r.GET("/employee-template", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.tmpl", employee.GetAll())
	})

	r.GET("/employee-template/:emloyeeID", func(ctx *gin.Context) {
		emloyeeID := ctx.Param("emloyeeID")

		if foundEmployee, ok := getEmployeeByID(ctx, emloyeeID); ok {
			ctx.HTML(http.StatusOK, "employee.tmpl", *foundEmployee)
		}
	})

	r.POST("/employee-template/:emloyeeID", func(ctx *gin.Context) {
		var timeoff employee.TimeOff
		err := ctx.ShouldBind(&timeoff)

		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		timeoff.Type = employee.TimeoffTypePTO
		timeoff.Status = employee.TimeoffStatusRequested

		emloyeeID := ctx.Param("emloyeeID")
		if foundEmployee, ok := getEmployeeByID(ctx, emloyeeID); ok {
			foundEmployee.TimeOff = append(foundEmployee.TimeOff, timeoff)
			ctx.Redirect(http.StatusFound, "/employee-template/"+emloyeeID)
		}
	})

	// JSON data
	g := r.Group("/api/json-employees", Benchmark)

	g.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, employee.GetAll())
	})

	g.GET("/:emloyeeID", func(ctx *gin.Context) {
		emloyeeID := ctx.Param("emloyeeID")

		if foundEmployee, ok := getEmployeeByID(ctx, emloyeeID); ok {
			ctx.JSON(http.StatusOK, *foundEmployee)
		}
	})

	g.POST("/:emloyeeID", func(ctx *gin.Context) {
		var timeoff employee.TimeOff
		err := ctx.ShouldBindJSON(&timeoff)

		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		timeoff.Type = employee.TimeoffTypePTO
		timeoff.Status = employee.TimeoffStatusRequested

		emloyeeID := ctx.Param("emloyeeID")
		if foundEmployee, ok := getEmployeeByID(ctx, emloyeeID); ok {
			foundEmployee.TimeOff = append(foundEmployee.TimeOff, timeoff)

			ctx.JSON(http.StatusOK, *foundEmployee)
		}
	})

	r.GET("/errors", func(ctx *gin.Context) {
		err := &gin.Error{
			Err:  errors.New("Something wrong"),
			Type: gin.ErrorTypeRender | gin.ErrorTypePublic,
			Meta: "This error was intentional",
		}

		ctx.Error(err)
	})
}

func getEmployeeByID(ctx *gin.Context, employeeID string) (*employee.Employee, bool) {
	employeeIDInt, err := strconv.Atoi(employeeID)

	if err != nil {
		ctx.AbortWithStatus((http.StatusNotFound))
		return nil, false
	}

	foundEmployee, err := employee.Get(employeeIDInt)
	if err != nil {
		ctx.AbortWithStatus((http.StatusInternalServerError))
		return nil, false
	}

	return foundEmployee, true
}

var Benchmark gin.HandlerFunc = func(ctx *gin.Context) {
	t := time.Now()

	ctx.Next()

	elapsed := time.Since(t)
	log.Print("Time to process:", elapsed)
}

var ErrorMiddleware gin.HandlerFunc = func(ctx *gin.Context) {
	ctx.Next()

	for _, err := range ctx.Errors {
		log.Print(map[string]any{
			"Err":      err.Error(),
			"Type":     err.Type,
			"Metadata": err.Meta,
		})
	}
}
