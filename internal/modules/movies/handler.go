package movies

import (
	"net/http"

	"github.com/boichique/movie-reviews/contracts"
	"github.com/boichique/movie-reviews/internal/config"
	"github.com/boichique/movie-reviews/internal/echox"
	"github.com/boichique/movie-reviews/internal/modules/genres"
	"github.com/boichique/movie-reviews/internal/modules/stars"
	"github.com/boichique/movie-reviews/internal/pagination"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/singleflight"
)

type Handler struct {
	service          *Service
	paginationConfig config.PaginationConfig
	reqGroup         singleflight.Group
}

func NewHandler(service *Service, paginationConfig config.PaginationConfig) *Handler {
	return &Handler{
		service:          service,
		paginationConfig: paginationConfig,
	}
}

func (h *Handler) Create(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.CreateMovieRequest](c)
	if err != nil {
		return err
	}
	movie := &MovieDetails{
		Movie: Movie{
			Title:       req.Title,
			ReleaseDate: req.ReleaseDate,
		},
		Description: req.Description,
	}
	for _, genreID := range req.GenresID {
		movie.Genres = append(movie.Genres, &genres.Genre{ID: genreID})
	}

	for _, creditID := range req.Cast {
		movie.Cast = append(
			movie.Cast, &stars.MovieCredit{
				Star: stars.Star{
					ID: creditID.StarID,
				},
				Role:    creditID.Role,
				Details: creditID.Details,
			},
		)
	}

	err = h.service.Create(c.Request().Context(), movie)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, movie)
}

func (h *Handler) GetMoviesPaginated(c echo.Context) error {
	res, err, _ := h.reqGroup.Do(c.Request().RequestURI, func() (any, error) {
		req, err := echox.BindAndValidate[contracts.GetMoviesPaginatedRequest](c)
		if err != nil {
			return nil, err
		}
		pagination.SetDefaults(&req.PaginatedRequest, h.paginationConfig)
		offset, limit := pagination.OffsetLimit(&req.PaginatedRequest)

		movies, total, err := h.service.GetMoviesPaginated(c.Request().Context(), req.SearchTerm, req.StarID, req.SortByRating, offset, limit)
		if err != nil {
			return nil, err
		}
		return pagination.Response(&req.PaginatedRequest, total, movies), nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetByID(c echo.Context) error {
	res, err, _ := h.reqGroup.Do(c.Request().RequestURI, func() (any, error) {
		req, err := echox.BindAndValidate[contracts.GetMovieRequest](c)
		if err != nil {
			return err, nil
		}
		return h.service.GetByID(c.Request().Context(), req.MovieID)
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) Update(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.UpdateMovieRequest](c)
	if err != nil {
		return err
	}
	movie := &MovieDetails{
		Movie: Movie{
			ID:          req.MovieID,
			Title:       req.Title,
			ReleaseDate: req.ReleaseDate,
		},
		Description: req.Description,
	}
	for _, genreID := range req.GenresID {
		movie.Genres = append(movie.Genres, &genres.Genre{ID: genreID})
	}

	for _, creditID := range req.Cast {
		movie.Cast = append(
			movie.Cast, &stars.MovieCredit{
				Star: stars.Star{
					ID: creditID.StarID,
				},
				Role:    creditID.Role,
				Details: creditID.Details,
			},
		)
	}

	if err = h.service.Update(c.Request().Context(), movie); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) Delete(c echo.Context) error {
	req, err := echox.BindAndValidate[contracts.DeleteMovieRequest](c)
	if err != nil {
		return err
	}

	if err = h.service.Delete(c.Request().Context(), req.MovieID); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
