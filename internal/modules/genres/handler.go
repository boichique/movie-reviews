package genres

import (
	"net/http"

	"github.com/boichique/movie-reviews/contracts"
	"github.com/boichique/movie-reviews/internal/echox"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.CreateGenreRequest](c)
	if err != nil {
		return err
	}

	genre, err := h.service.Create(c.Request().Context(), req.Name)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, genre)
}

func (h *Handler) GetGenres(c echo.Context) error {
	genres, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, genres)
}

func (h *Handler) GetByID(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.GetOrDeleteGenreRequest](c)
	if err != nil {
		return err
	}

	genre, err := h.service.GetByID(c.Request().Context(), req.GenreID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, genre)
}

func (h *Handler) UpdateName(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.UpdateGenreRequest](c)
	if err != nil {
		return err
	}

	return h.service.Update(c.Request().Context(), req.GenreID, req.Name)
}

func (h *Handler) Delete(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.GetOrDeleteGenreRequest](c)
	if err != nil {
		return err
	}

	return h.service.Delete(c.Request().Context(), req.GenreID)
}
