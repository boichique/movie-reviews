package auth

import (
	"net/http"

	"github.com/boichique/movie-reviews/contracts"
	"github.com/boichique/movie-reviews/internal/echox"
	"github.com/boichique/movie-reviews/internal/modules/users"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	authService *Service
}

func NewHandler(authService *Service) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) Register(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.RegisterUserRequest](c)
	if err != nil {
		return err
	}

	user := &users.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     users.UserRole,
	}

	if err := h.authService.Register(c.Request().Context(), user, req.Password); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *Handler) Login(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.LoginUserRequest](c)
	if err != nil {
		return err
	}

	accessToken, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, LoginResponse{AccessToken: accessToken})
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}
