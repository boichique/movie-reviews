package echox

import (
	"errors"
	"log"
	"net/http"

	"github.com/boichique/movie-reviews/internal/apperrors"
	"github.com/labstack/echo/v4"
)

type HttpError struct {
	Message    string `json:"message"`
	IncidentID string `json:"incidentId,omitempty"`
}

func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var appError *apperrors.Error
	if !errors.As(err, &appError) {
		appError = apperrors.InternalWithoutStackTrace(err)
	}

	httpError := HttpError{
		Message:    appError.SafeError(),
		IncidentID: appError.IncidentID,
	}

	if appError.Code == apperrors.InternalCode {
		log.Printf(
			"[ERROR] %s %s : %s\nincidentID: %s\nstacktrace: %s",
			c.Request().Method,
			c.Request().RequestURI,
			err.Error(),
			appError.IncidentID,
			appError.StackTrace,
		)
	} else {
		log.Printf(
			"[WARNING] %s %s : %s",
			c.Request().Method,
			c.Request().RequestURI,
			err.Error(),
		)
	}

	if err = c.JSON(toHTTPStatus(appError.Code), httpError); err != nil {
		c.Logger().Error(err)
	}
}

func toHTTPStatus(code apperrors.Code) int {
	switch code {
	case apperrors.InternalCode:
		return http.StatusInternalServerError
	case apperrors.BadRequestCode:
		return http.StatusBadRequest
	case apperrors.NotFoundCode:
		return http.StatusNotFound
	case apperrors.AlreadyExistsCode:
		return http.StatusConflict
	case apperrors.UnauthorizedCode:
		return http.StatusUnauthorized
	case apperrors.ForbiddenCode:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
