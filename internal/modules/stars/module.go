package stars

import (
	"github.com/boichique/movie-reviews/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	Handler    *Handler
	Service    *Service
	Repository *Repository
}

func NewModule(db *pgxpool.Pool, pagintaionConfig config.PaginationConfig) *Module {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service, pagintaionConfig)

	return &Module{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}
