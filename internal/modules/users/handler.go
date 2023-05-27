package users

import (
	"net/http"

	"github.com/boichique/movie-reviews/internal/apperrors"
	"github.com/boichique/movie-reviews/internal/echox"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h Handler) Get(c echo.Context) error {
	req, err := echox.BindAndValidate[UserIDRequest](c)
	if err != nil {
		return err
	}

	user, err := h.service.GetExistingUserByID(c.Request().Context(), req.UserID)
	if err != nil {
		if apperrors.Is(err, apperrors.InternalCode) {
			return apperrors.Internal(err)
		} else {
			return apperrors.BadRequest(err)
		}
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) UpdateBio(c echo.Context) error {
	req, err := echox.BindAndValidate[UpdateBioRequest](c)
	if err != nil {
		return err
	}

	return h.service.UpdateBio(c.Request().Context(), req.UserID, req.Bio)
}

func (h *Handler) UpdateRole(c echo.Context) error {
	req, err := echox.BindAndValidate[UpdateRoleRequest](c)
	if err != nil {
		return err
	}

	return h.service.UpdateRole(c.Request().Context(), req.UserID, req.Role)
}

func (h *Handler) Delete(c echo.Context) error {
	req, err := echox.BindAndValidate[UserIDRequest](c)
	if err != nil {
		return err
	}

	return h.service.DeleteUser(c.Request().Context(), req.UserID)
}

type UserIDRequest struct {
	UserID int `param:"userID"`
}

type UpdateBioRequest struct {
	UserID int    `param:"userID" validate:"nonzero"`
	Bio    string `json:"bio"`
}

type UpdateRoleRequest struct {
	UserID int    `param:"userID" validate:"nonzero"`
	Role   string `param:"role" validate:"role"`
}
