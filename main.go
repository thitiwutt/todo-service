package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/thitiwutt/todoapi/auth"
	"github.com/thitiwutt/todoapi/todo"
	"golang.org/x/time/rate"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NOTE - Idflags
var (
	buildCommit = "dev"
	buildTime   = time.Now().String()
)

func main() {
	// SECTION - Liveness probe
	_, err := os.Create("/tmp/live")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove("/tmp/live")

	// SECTION - Load env
	err = godotenv.Load("local.env")
	if err != nil {
		fmt.Println("NOT HAVE LOCAL ENV")
	}

	db, err := gorm.Open(mysql.Open(os.Getenv("DB_CONN")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&todo.Todo{})

	// Default returns an Engine instance with the Logger and Recovery middleware already attached.
	r := gin.Default()

	// SECTION - CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:8080",
	}
	config.AllowHeaders = []string{
		"Origin",
		"Authorization",
		"TransactionID",
	}
	r.Use(cors.New(config))

	// SECTION - Readiness probe
	r.GET("/health", func(ctx *gin.Context) {
		ctx.Status(200)
	})
	// SECTION - rate limit
	r.GET("/limit", limitedHandler)
	// SECTION - Idflags
	r.GET("/x", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"buildCommit": buildCommit,
			"buildTime":   buildTime,
		})
	})

	signature := os.Getenv("SIGN")
	r.GET("/tokenz", auth.AccessToken(signature))

	// group of middleware
	protected := r.Group("", auth.Protect(signature))

	todoHandler := todo.NewTodoHandler(db)
	protected.POST("/todo", todoHandler.NewTask)
	protected.GET("/todo", todoHandler.ListTodos)
	protected.DELETE("/todo/:id", todoHandler.DeleteTodo)

	// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
	// r.Run()

	// SECTION - gracefully shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s := &http.Server{
		Addr:           ":" + os.Getenv("PORT"),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen %s \n", err)
		}
	}()

	<-ctx.Done()
	stop()
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}

}

// SECTION - rate limit
var limiter = rate.NewLimiter(5, 5)

func limitedHandler(ctx *gin.Context) {
	if !limiter.Allow() {
		ctx.AbortWithStatus(http.StatusTooManyRequests)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
