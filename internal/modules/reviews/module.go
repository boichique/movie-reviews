package reviews

import (
	"github.com/boichique/movie-reviews/internal/config"
	"github.com/boichique/movie-reviews/internal/modules/movies"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	Handler    *Handler
	Service    *Service
	Repository *Repository
}

func NewModule(db *pgxpool.Pool, moviesModule *movies.Module, paginationConfig config.PaginationConfig) *Module {
	repo := NewRepository(db, moviesModule.Repository)
	service := NewService(repo)
	handler := NewHandler(service, paginationConfig)

	return &Module{
		Handler:    handler,
		Service:    service,
		Repository: repo,
	}
}
