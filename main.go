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

	"github.com/boichique/movie-reviews/internal/apperrors"
	"github.com/boichique/movie-reviews/internal/config"
	"github.com/boichique/movie-reviews/internal/echox"
	"github.com/boichique/movie-reviews/internal/jwt"
	"github.com/boichique/movie-reviews/internal/modules/auth"
	"github.com/boichique/movie-reviews/internal/modules/users"
	"github.com/boichique/movie-reviews/internal/validation"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/validator.v2"
)

const (
	dbConnectTimeout     = 10 * time.Second
	adminCreationTimeout = 5 * time.Second
	gracefulTimeout      = 10 * time.Second
)

func main() {
	e := echo.New()
	validation.SetupValidators()

	cfg, err := config.NewConfig()
	failOnError(err, "parse config")

	db, err := getDB(context.Background(), cfg.DBUrl)
	failOnError(err, "connect to database")

	e.HTTPErrorHandler = echox.ErrorHandler
	jwtService := jwt.NewService(cfg.Jwt.Secret, cfg.Jwt.AccessExpiration)
	usersModule := users.NewModule(db)
	authModule := auth.NewModule(usersModule.Service, jwtService)

	authMiddleware := jwt.NewAuthMiddleware(cfg.Jwt.Secret)

	_ = CreateAdmin(cfg.Admin, authModule.Service)

	e.Use(middleware.Recover())
	api := e.Group("/api")
	api.Use(authMiddleware)

	// auth group
	api.POST("/auth/register", authModule.Handler.Register)
	api.POST("/auth/login", authModule.Handler.Login)

	// users group
	api.GET("/users/:userID", usersModule.Handler.Get)
	api.PUT("/users/:userID", usersModule.Handler.UpdateBio, auth.Self)
	api.PUT("/users/:userID/role/:role", usersModule.Handler.UpdateRole, auth.Admin)
	api.DELETE("/users/:userID", usersModule.Handler.Delete, auth.Self)

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

func CreateAdmin(cfg config.AdminConfig, authService *auth.Service) error {
	if !cfg.AdminIsSet() {
		return nil
	}

	if err := validator.Validate(cfg); err != nil {
		return fmt.Errorf("validate admin config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), adminCreationTimeout)
	defer cancel()

	err := authService.Register(ctx, &users.User{
		Username: cfg.AdminName,
		Email:    cfg.AdminEmail,
		Role:     users.AdminRole,
	}, cfg.AdminPassword)

	switch {
	case apperrors.Is(err, apperrors.InternalCode):
		return fmt.Errorf("register admin using config: %w", err)
	default:
		return err
	}
}
