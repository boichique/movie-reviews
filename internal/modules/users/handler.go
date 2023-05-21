package users

import "github.com/labstack/echo/v4"

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h Handler) GetUsers(c echo.Context) error {
	return c.String(200, "not implemented")
}
