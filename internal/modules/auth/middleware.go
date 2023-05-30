package auth

import (
	"github.com/boichique/movie-reviews/internal/apperrors"
	"github.com/boichique/movie-reviews/internal/jwt"
	"github.com/boichique/movie-reviews/internal/modules/users"
	"github.com/labstack/echo/v4"
)

var (
	errForbidden    = apperrors.Forbidden("not enough permissions")
	errUnauthorized = apperrors.Unauthorized("unauthorized user")
)

func Self(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Param("userID")

		claims := jwt.GetClaims(c)
		if claims == nil {
			return errUnauthorized
		}

		if claims.Role == users.AdminRole || claims.Subject == userID {
			return next(c)
		}

		return errForbidden
	}
}

func Editor(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims := jwt.GetClaims(c)
		if claims == nil {
			return errUnauthorized
		}

		if claims.Role == users.AdminRole || claims.Role == users.EditorRole {
			return next(c)
		}

		return errForbidden
	}
}

func Admin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims := jwt.GetClaims(c)
		if claims == nil {
			return errUnauthorized
		}

		if claims.Role == users.AdminRole {
			return next(c)
		}

		return errForbidden
	}
}
