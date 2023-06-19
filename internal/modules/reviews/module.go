package reviews

import (
	"github.com/boichique/movie-reviews/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	Handler    *Handler
	Service    *Service
	Repository *Repository
}

func NewModule(db *pgxpool.Pool, paginationConfig config.PaginationConfig) *Module {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service, paginationConfig)

	return &Module{
		Handler:    handler,
		Service:    service,
		Repository: repo,
	}
}
