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

	"github.com/cloudmachinery/movie-reviews/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

const (
	dbConnectTimeout = 10 * time.Second
)

func main() {
	e := echo.New()

	cfg, err := config.NewConfig()
	failOnError(err, "parse config")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	db, err := getDB(context.Background(), cfg.DBUrl)
	failOnError(err, "connect to database")

	err = db.Ping(context.Background())
	failOnError(err, "ping database")

	go func() {
		sig := <-sigChan
		log.Printf("\nreceived signal %s. Shutting down...\n", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
