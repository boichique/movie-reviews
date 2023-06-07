package stars

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
	req, err := echox.BindAndValidate[contracts.CreateStarRequest](c)
	if err != nil {
		return err
	}

	star := &Star{
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		LastName:   req.LastName,
		BirthDate:  req.BirthDate,
		BirthPlace: req.BirthPlace,
		DeathDate:  req.DeathDate,
		Bio:        req.Bio,
	}

	err = h.service.Create(c.Request().Context(), star)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, star)
}

func (h *Handler) GetStars(c echo.Context) error {
	stars, err := h.service.GetStars(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, stars)
}

func (h *Handler) GetByID(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.GetOrDeleteStarRequest](c)
	if err != nil {
		return err
	}

	star, err := h.service.GetByID(c.Request().Context(), req.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, star)
}

func (h *Handler) Delete(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.GetOrDeleteStarRequest](c)
	if err != nil {
		return err
	}

	if err = h.service.Delete(c.Request().Context(), req.ID); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
