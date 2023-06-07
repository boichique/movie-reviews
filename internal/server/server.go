package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/boichique/movie-reviews/internal/apperrors"
	"github.com/boichique/movie-reviews/internal/config"
	"github.com/boichique/movie-reviews/internal/echox"
	"github.com/boichique/movie-reviews/internal/jwt"
	"github.com/boichique/movie-reviews/internal/log"
	"github.com/boichique/movie-reviews/internal/modules/auth"
	"github.com/boichique/movie-reviews/internal/modules/genres"
	"github.com/boichique/movie-reviews/internal/modules/users"
	"github.com/boichique/movie-reviews/internal/validation"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/exp/slog"
	"gopkg.in/validator.v2"
)

const (
	dbConnectTimeout     = 10 * time.Second
	adminCreationTimeout = 5 * time.Second
)

type Server struct {
	e       *echo.Echo
	cfg     *config.Config
	closers []func() error
}

func New(ctx context.Context, cfg *config.Config) (*Server, error) {
	logger, err := log.SetupLogger(cfg.Local, cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("setup logger: %w", err)
	}

	slog.SetDefault(logger)
	validation.SetupValidators()

	var closers []func() error
	db, err := getDB(context.Background(), cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	closers = append(closers, func() error {
		db.Close()
		return nil
	})

	e := echo.New()
	e.HTTPErrorHandler = echox.ErrorHandler
	jwtService := jwt.NewService(cfg.Jwt.Secret, cfg.Jwt.AccessExpiration)
	usersModule := users.NewModule(db)
	authModule := auth.NewModule(usersModule.Service, jwtService)
	authMiddleware := jwt.NewAuthMiddleware(cfg.Jwt.Secret)
	genreModule := genres.NewModule(db)

	if err = createAdmin(cfg.Admin, authModule.Service); err != nil {
		return nil, withClosers(closers, fmt.Errorf("create admin: %w", err))
	}

	e.Use(middleware.Recover())
	e.HideBanner = true
	e.HidePort = true

	api := e.Group("/api")
	api.Use(authMiddleware)
	api.Use(echox.Logger)

	// auth group
	api.POST("/auth/register", authModule.Handler.Register)
	api.POST("/auth/login", authModule.Handler.Login)

	// users group
	api.GET("/users/:userID", usersModule.Handler.GetByID)
	api.GET("/users/username/:username", usersModule.Handler.GetByUsername)
	api.PUT("/users/:userID", usersModule.Handler.UpdateBio, auth.Self)
	api.PUT("/users/:userID/role/:role", usersModule.Handler.UpdateRole, auth.Admin)
	api.DELETE("/users/:userID", usersModule.Handler.Delete, auth.Self)

	// genres group
	api.POST("/genres", genreModule.Handler.Create, auth.Editor)
	api.GET("/genres", genreModule.Handler.GetGenres)
	api.GET("/genres/:genreID", genreModule.Handler.GetByID)
	api.PUT("/genres/:genreID", genreModule.Handler.Update, auth.Editor)
	api.DELETE("/genres/:genreID", genreModule.Handler.Delete, auth.Editor)

	return &Server{
		e:       e,
		cfg:     cfg,
		closers: closers,
	}, nil
}

func (s *Server) Start() error {
	port := s.cfg.Port
	slog.Info(
		"Server starting",
		"port", port,
	)

	return s.e.Start(fmt.Sprintf(":%d", port))
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}

func (s *Server) Close() error {
	return withClosers(s.closers, nil)
}

func (s *Server) Port() (int, error) {
	listener := s.e.Listener
	if listener == nil {
		return 0, fmt.Errorf("server is not started")
	}

	addr := listener.Addr()
	if addr == nil {
		return 0, fmt.Errorf("server is not started")
	}

	return addr.(*net.TCPAddr).Port, nil
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

func createAdmin(cfg config.AdminConfig, authService *auth.Service) error {
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
		return fmt.Errorf("register admin: %w", err)
	case err != nil:
		slog.Info(
			"admin user already created",
			"username", cfg.AdminName,
			"email", cfg.AdminEmail,
		)
	}

	return nil
}

func withClosers(closers []func() error, err error) error {
	errs := []error{err}

	for i := len(closers) - 1; i >= 0; i-- {
		if err := closers[i](); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
