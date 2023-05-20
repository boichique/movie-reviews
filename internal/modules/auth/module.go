package auth

import (
	"github.com/boichique/movie-reviews/internal/jwt"
	"github.com/boichique/movie-reviews/internal/modules/users"
)

type Module struct {
	Handler    *Handler
	Service    *Service
	Repository *Repository
}

func NewModule(userService *users.Service, jwtService *jwt.Service) *Module {
	repository := NewRepository()
	service := NewService(userService, jwtService)
	handler := NewHandler(service)

	return &Module{
		Handler:    handler,
		Service:    service,
		Repository: repository,
	}
}
