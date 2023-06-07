package pagination

import (
	"github.com/boichique/movie-reviews/contracts"
	"github.com/boichique/movie-reviews/internal/config"
)

func SetDefaults(r *contracts.PaginatedRequest, cfg config.PaginationConfig) {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Size == 0 {
		r.Size = cfg.DefaultSize
	}

	if r.Size > cfg.MaxSize {
		r.Size = cfg.MaxSize
	}
}

func OffsetLimit(r *contracts.PaginatedRequest) (int, int) {
	offset := (r.Page - 1) * r.Size
	limit := r.Size
	return offset, limit
}

func Response[T any](r *contracts.PaginatedRequest, total int, items []*T) *contracts.PaginatedResponse[T] {
	return &contracts.PaginatedResponse[T]{
		Page:  r.Page,
		Size:  r.Size,
		Total: total,
		Items: items,
	}
}
