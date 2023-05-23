package users

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h Handler) Get(c echo.Context) error {
	var req UserIDRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := h.service.GetExistingUserByID(c.Request().Context(), req.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) Update(c echo.Context) error {
	var req PutRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return h.service.UpdateUser(c.Request().Context(), req.UserID, req.Bio)
}

func (h *Handler) Delete(c echo.Context) error {
	var req UserIDRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return h.service.DeleteUser(c.Request().Context(), req.UserID)
}

type UserIDRequest struct {
	UserID int `param:"userID"`
}

type PutRequest struct {
	UserID int    `param:"userID" validate:"nonzero"`
	Bio    string `json:"bio"`
}
