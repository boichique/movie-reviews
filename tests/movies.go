package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/contracts"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
)

var (
	StarTrek *contracts.MovieDetails
	StarWars *contracts.MovieDetails
)

func moviesAPIChecks(t *testing.T, c *client.Client) {
	t.Run("movies.CreateMovie: success", func(t *testing.T) {
		cases := []struct {
			req  *contracts.CreateMovieRequest
			addr **contracts.MovieDetails
		}{
			{
				req: &contracts.CreateMovieRequest{
					Title:       "Star Wars: Episode IV - A New Hope",
					ReleaseDate: time.Date(1977, time.May, 25, 0, 0, 0, 0, time.UTC),
					Description: "Fun movie about space wizards",
					GenresID:    []int{Adventure.ID, Drama.ID},
					Cast: []*contracts.MovieCreditInfo{
						{
							StarID: GeorgeLucas.ID,
							Role:   "director",
						},
						{
							StarID:  EvanMcGregor.ID,
							Role:    "actor",
							Details: ptr("Obi-Wan Kenobi"),
						},
					},
				},
				addr: &StarWars,
			},
			{
				req: &contracts.CreateMovieRequest{
					Title:       "Star Trek: The Motion Picture",
					ReleaseDate: time.Date(1979, time.December, 7, 0, 0, 0, 0, time.UTC),
					Description: "When an alien spacecraft of enormous power is spotted approaching Earth, " +
						"Admiral James T. Kirk resumes command of the overhauled USS Enterprise in order to " +
						"intercept it.",
					GenresID: []int{Adventure.ID},
					Cast: []*contracts.MovieCreditInfo{
						{
							StarID:  WilliamShatner.ID,
							Role:    "actor",
							Details: ptr("Admiral James T. Kirk"),
						},
					},
				},
				addr: &StarTrek,
			},
		}

		for _, cc := range cases {
			movie, err := c.CreateMovie(contracts.NewAuthenticated(cc.req, johnDoeToken))
			require.NoError(t, err)

			*cc.addr = movie
			require.NotEmpty(t, movie.ID)
			require.Len(t, movie.Genres, len(cc.req.GenresID))
			for i, genreID := range cc.req.GenresID {
				require.Equal(t, genreID, movie.Genres[i].ID)
				require.NotNil(t, movie.Genres[i].Name)
			}
			for i, credit := range cc.req.Cast {
				require.Equal(t, credit.StarID, movie.Cast[i].Star.ID)
				require.NotNil(t, movie.Cast[i].Star.FirstName)
				require.NotNil(t, movie.Cast[i].Star.LastName)
				require.Equal(t, credit.Role, movie.Cast[i].Role)
				require.Equal(t, credit.Details, movie.Cast[i].Details)
			}
		}
	})

	t.Run("movies.GetMovie: success", func(t *testing.T) {
		for _, movie := range []*contracts.MovieDetails{StarWars, StarTrek} {
			m, err := c.GetMovie(movie.ID)
			require.NoError(t, err)

			require.Equal(t, movie, m)
		}
	})

	t.Run("movies.GetMovies: success", func(t *testing.T) {
		req := &contracts.GetMoviesPaginatedRequest{}
		res, err := c.GetMovies(req)
		require.NoError(t, err)

		require.Equal(t, 2, res.Total)
		require.Equal(t, 1, res.Page)
		require.Equal(t, testPaginationSize, res.Size)
		require.Equal(t, []*contracts.Movie{&StarWars.Movie, &StarTrek.Movie}, res.Items)
	})

	t.Run("movies.UpdateMovie: success", func(t *testing.T) {
		req := &contracts.UpdateMovieRequest{
			MovieID:     StarWars.ID,
			Title:       StarWars.Title,
			ReleaseDate: StarWars.ReleaseDate,
			Description: "Luke Skywalker joins forces with a Jedi Knight, a cocky pilot, a Wookiee and " +
				"two droids to save the galaxy from the Empire's world-destroying battle station, " +
				"while also attempting to rescue Princess Leia from the mysterious Darth Vader.",
			GenresID: []int{Adventure.ID, Action.ID, SciFi.ID},
			Cast: []*contracts.MovieCreditInfo{
				{
					StarID: GeorgeLucas.ID,
					Role:   "director",
				},
				{
					StarID:  MarkHamill.ID,
					Role:    "actor",
					Details: ptr("Luke Skywalker"),
				},
			},
		}

		err := c.UpdateMovie(contracts.NewAuthenticated(req, johnDoeToken))
		require.NoError(t, err)

		// Concurrent update should fail
		err = c.UpdateMovie(contracts.NewAuthenticated(req, johnDoeToken))
		requireVersionMismatchError(t, err, "movie", "id", req.MovieID, req.Version)

		StarWars = getMovie(t, c, StarWars.ID)
		require.Equal(t, req.Description, StarWars.Description)
		require.Equal(t, []*contracts.Genre{Adventure, Action, SciFi}, StarWars.Genres)
		require.Len(t, StarWars.Cast, len(req.Cast))
		require.Equal(t, 1, StarWars.Version)
		for i, credit := range req.Cast {
			require.Equal(t, credit.StarID, StarWars.Cast[i].Star.ID)
			require.NotNil(t, StarWars.Cast[i].Star.FirstName)
			require.NotNil(t, StarWars.Cast[i].Star.LastName)
			require.Equal(t, credit.Role, StarWars.Cast[i].Role)
			require.Equal(t, credit.Details, StarWars.Cast[i].Details)
		}
	})

	t.Run("movies.DeleteMovie: success", func(t *testing.T) {
		movie := createRandomMovie(t, c)
		err := c.DeleteMovie(contracts.NewAuthenticated(&contracts.DeleteMovieRequest{MovieID: movie.ID}, johnDoeToken))
		require.NoError(t, err)

		movie = getMovie(t, c, movie.ID)
		require.Nil(t, movie)
	})

	t.Run("stars.GetStars: from Star Wars", func(t *testing.T) {
		req := &contracts.GetStarsPaginatedRequest{
			MovieID: ptr(StarWars.ID),
		}
		res, err := c.GetStars(req)
		require.NoError(t, err)

		require.Equal(t, 2, res.Total)
		require.Equal(t, 1, res.Page)
		require.Equal(t, testPaginationSize, res.Size)
		require.Equal(t, []*contracts.Star{&GeorgeLucas.Star, &MarkHamill.Star}, res.Items)
	})

	t.Run("movies.GetMovies: of George Lucas", func(t *testing.T) {
		req := &contracts.GetMoviesPaginatedRequest{
			StarID: ptr(GeorgeLucas.ID),
		}
		res, err := c.GetMovies(req)
		require.NoError(t, err)

		require.Equal(t, 1, res.Total)
		require.Equal(t, 1, res.Page)
		require.Equal(t, testPaginationSize, res.Size)
		require.Equal(t, []*contracts.Movie{&StarWars.Movie}, res.Items)
	})
}

func getMovie(t *testing.T, c *client.Client, id int) *contracts.MovieDetails {
	m, err := c.GetMovie(id)
	if err != nil {
		cerr, ok := err.(*client.Error)
		require.True(t, ok)
		require.Equal(t, http.StatusNotFound, cerr.Code)
		return nil
	}

	return m
}

func createRandomMovie(t *testing.T, c *client.Client) *contracts.MovieDetails {
	r := rand.Intn(10000)

	req := &contracts.CreateMovieRequest{
		Title:       fmt.Sprintf("Movie #%d", r),
		ReleaseDate: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
		Description: fmt.Sprintf("Description for movie #%d", r),
		GenresID:    []int{Action.ID},
	}
	movie, err := c.CreateMovie(contracts.NewAuthenticated(req, johnDoeToken))
	require.NoError(t, err)

	return movie
}
