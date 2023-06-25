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
	"github.com/boichique/movie-reviews/internal/modules/movies"
	"github.com/boichique/movie-reviews/internal/modules/reviews"
	"github.com/boichique/movie-reviews/internal/modules/stars"
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
	db, err := getDB(ctx, cfg.DBUrl)
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
	starsModule := stars.NewModule(db, cfg.Pagination)
	moviesModule := movies.NewModule(db, genreModule, starsModule, cfg.Pagination)
	reviewsModule := reviews.NewModule(db, cfg.Pagination)

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
	api.PUT("/genres/:genreID", genreModule.Handler.UpdateName, auth.Editor)
	api.DELETE("/genres/:genreID", genreModule.Handler.Delete, auth.Editor)

	// stars group
	api.POST("/stars", starsModule.Handler.Create, auth.Editor)
	api.GET("/stars", starsModule.Handler.GetStarsPaginated)
	api.GET("/stars/:starID", starsModule.Handler.GetByID)
	api.PUT("/stars/:starID", starsModule.Handler.Update, auth.Editor)
	api.DELETE("/stars/:starID", starsModule.Handler.Delete, auth.Editor)

	// movies group
	api.POST("/movies", moviesModule.Handler.Create, auth.Editor)
	api.GET("/movies", moviesModule.Handler.GetMoviesPaginated)
	api.GET("/movies/:movieID", moviesModule.Handler.GetByID)
	api.PUT("/movies/:movieID", moviesModule.Handler.Update, auth.Editor)
	api.DELETE("/movies/:movieID", moviesModule.Handler.Delete, auth.Editor)

	// reviews group
	api.POST("/users/:userID/reviews", reviewsModule.Handler.Create, auth.Self)
	api.GET("/reviews", reviewsModule.Handler.GetReviewsPaginated)
	api.GET("/reviews/:reviewID", reviewsModule.Handler.GetByID)
	api.PUT("/users/:userID/reviews/:reviewID", reviewsModule.Handler.Update, auth.Self)
	api.DELETE("/users/:userID/reviews/:reviewID", reviewsModule.Handler.Delete, auth.Self)

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
