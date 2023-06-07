package users

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

func (h Handler) GetByID(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.GetOrDeleteUserRequest](c)
	if err != nil {
		return err
	}

	user, err := h.service.GetExistingUserByID(c.Request().Context(), req.UserID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}

func (h Handler) GetByUsername(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.GetUserByUsernameRequest](c)
	if err != nil {
		return err
	}

	user, err := h.service.GetExistingUserByUsername(c.Request().Context(), req.Username)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) UpdateBio(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.UpdateUserBioRequest](c)
	if err != nil {
		return err
	}

	return h.service.UpdateBio(c.Request().Context(), req.UserID, *req.Bio)
}

func (h *Handler) UpdateRole(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.UpdateUserRoleRequest](c)
	if err != nil {
		return err
	}

	return h.service.UpdateRole(c.Request().Context(), req.UserID, req.Role)
}

func (h *Handler) Delete(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.GetOrDeleteUserRequest](c)
	if err != nil {
		return err
	}

	return h.service.DeleteUser(c.Request().Context(), req.UserID)
}
