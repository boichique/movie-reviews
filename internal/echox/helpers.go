package echox

import (
	"github.com/boichique/movie-reviews/internal/apperrors"
	"github.com/labstack/echo/v4"
	"gopkg.in/validator.v2"
)

func BindAndValidate[T any](c echo.Context) (*T, error) {
	req := new(T)
	if err := c.Bind(req); err != nil {
		return nil, apperrors.BadRequestHidden(err, "invalid request")
	}

	if err := validator.Validate(req); err != nil {
		return nil, apperrors.BadRequest(err)
	}

	return req, nil
}
