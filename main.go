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

	"github.com/boichique/movie-reviews/internal/config"
	"github.com/boichique/movie-reviews/internal/jwt"
	"github.com/boichique/movie-reviews/internal/modules/auth"
	"github.com/boichique/movie-reviews/internal/modules/users"
	"github.com/boichique/movie-reviews/internal/validation"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

const (
	dbConnectTimeout = 10 * time.Second
	gracefulTimeout  = 10 * time.Second
)

func main() {
	e := echo.New()
	validation.SetupValidators()

	cfg, err := config.NewConfig()
	failOnError(err, "parse config")

	db, err := getDB(context.Background(), cfg.DBUrl)
	failOnError(err, "connect to database")

	jwtService := jwt.NewService(cfg.Jwt.Secret, cfg.Jwt.AccessExpiration)
	usersModule := users.NewModule(db)
	authModule := auth.NewModule(usersModule.Service, jwtService)

	authMiddleware := jwt.NewAuthMiddleware(cfg.Jwt.Secret)

	apiAuth := e.Group("/api/auth")
	apiAuth.POST("/register", authModule.Handler.Register)
	apiAuth.POST("/login", authModule.Handler.Login)

	apiUsers := e.Group("/api/users")
	apiUsers.GET("/:userID", usersModule.Handler.Get)
	apiUsers.PUT("/:userID", usersModule.Handler.Update, authMiddleware, auth.Self)
	apiUsers.DELETE("/:userID", usersModule.Handler.Delete, authMiddleware, auth.Self)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT)
		<-sigCh
		log.Println("received interrupt signal. Shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		if err := e.Shutdown(ctx); err != nil {
			log.Fatalf("shutdown server: %s", err)
		}
	}()

	err = e.Start(fmt.Sprintf(":%d", cfg.Port))
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func getDB(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, dbConnectTimeout)
	defer cancel()

	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}

	return db, nil
}

func failOnError(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %s", message, err)
	}
}
