package integrationTests

import (
	"net/http"
	"testing"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/internal/apperrors"
	"github.com/stretchr/testify/require"
)

func requireNotFoundError(t *testing.T, err error, subject, key string, value any) {
	msg := apperrors.NotFound(subject, key, value).Error()
	requireAPIError(t, err, http.StatusNotFound, msg)
}

func requireAlreadyExistsError(t *testing.T, err error, subject, key string, value any) {
	msg := apperrors.AlreadyExists(subject, key, value).Error()
	requireAPIError(t, err, http.StatusConflict, msg)
}

func requireUnauthorizedError(t *testing.T, err error, msg string) {
	requireAPIError(t, err, http.StatusUnauthorized, msg)
}

func requireForbiddenError(t *testing.T, err error, msg string) {
	requireAPIError(t, err, http.StatusForbidden, msg)
}

func requireBadRequestError(t *testing.T, err error, msg string) {
	requireAPIError(t, err, http.StatusBadRequest, msg)
}

func requireAPIError(t *testing.T, err error, statusCode int, msg string) {
	cerr, ok := err.(*client.Error)
	require.True(t, ok, "expected client.Error")
	require.Equal(t, statusCode, cerr.Code)
	require.Contains(t, cerr.Message, msg)
}
