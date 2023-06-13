package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/contracts"
	"github.com/stretchr/testify/require"
)

func moviesAPIChecks(t *testing.T, c *client.Client) {
	var starWars, harryPotter, lordOfTheRing *contracts.MovieDetails
	t.Run("movies.CreateMovie: success", func(t *testing.T) {
		cases := []struct {
			req  *contracts.CreateMovieRequest
			addr **contracts.MovieDetails
		}{
			{
				req: &contracts.CreateMovieRequest{
					Title:       "Star Wars",
					Description: "Star Wars is an American epic space opera",
					ReleaseDate: time.Date(1977, time.May, 25, 0, 0, 0, 0, time.UTC),
				},
				addr: &starWars,
			},
			{
				req: &contracts.CreateMovieRequest{
					Title: "Harry Poster and the Philosopher's Stone",
					Description: "is a 2001 fantasy film directed by Chris Columbus and produced by David Heyman," +
						" from a screenplay by Steve Kloves, based on the 1997 novel of the same name by J. K. Rowling." +
						" It is the first installment in the Harry Potter film series. ",
					ReleaseDate: time.Date(2001, time.November, 4, 0, 0, 0, 0, time.UTC),
				},
				addr: &harryPotter,
			},
			{
				req: &contracts.CreateMovieRequest{
					Title: "The Lord of the Rings. The Fellowship of the Ring",
					Description: "The Lord of the Rings is a series of three epic fantasy adventure films directed by Peter Jackson," +
						" based on the novel The Lord of the Rings by J. R. R. Tolkien",
					ReleaseDate: time.Date(2001, time.December, 10, 0, 0, 0, 0, time.UTC),
				},
				addr: &lordOfTheRing,
			},
		}

		for _, cc := range cases {

			movie, err := c.CreateMovie(contracts.NewAuthenticated(cc.req, johnDoeToken))
			require.NoError(t, err)

			*cc.addr = movie
			require.NotEmpty(t, movie.ID)
			require.NotEmpty(t, movie.CreatedAt)
		}
	})

	t.Run("movies.GetMovie: success", func(t *testing.T) {
		for _, movie := range []*contracts.MovieDetails{starWars, harryPotter, lordOfTheRing} {
			s, err := c.GetMovie(movie.ID)
			require.NoError(t, err)
			require.Equal(t, movie, s)
		}
	})

	t.Run("movies.GetMovie: not found", func(t *testing.T) {
		nonExistingID := 1000
		_, err := c.GetMovie(nonExistingID)
		requireNotFoundError(t, err, "movie", "id", nonExistingID)
	})

	t.Run("movies.UpdateMovie: success", func(t *testing.T) {
		req := &contracts.UpdateMovieRequest{
			ID:    harryPotter.ID,
			Title: "Harry Potter and the Philosopher's Stone",
			Description: "is a 2001 fantasy film directed by Chris Columbus and produced by David Heyman," +
				" from a screenplay by Steve Kloves, based on the 1997 novel of the same name by J. K. Rowling." +
				" It is the first installment in the Harry Potter film series. ",
			ReleaseDate: time.Date(2001, time.November, 4, 0, 0, 0, 0, time.UTC),
		}
		err := c.UpdateMovie(contracts.NewAuthenticated(req, johnDoeToken))
		require.NoError(t, err)

		harryPotter = getMovie(t, c, harryPotter.ID)
		require.Equal(t, req.Title, harryPotter.Title)
	})

	t.Run("movies.UpdateMovie: not found", func(t *testing.T) {
		nonExistingID := 1000
		req := &contracts.UpdateMovieRequest{
			ID:          nonExistingID,
			Title:       "...",
			Description: "...",
			ReleaseDate: time.Date(2000, time.May, 1, 0, 0, 0, 0, time.UTC),
		}
		err := c.UpdateMovie(contracts.NewAuthenticated(req, johnDoeToken))
		requireNotFoundError(t, err, "movie", "id", nonExistingID)
	})
	t.Run("movies.GetMovies: success", func(t *testing.T) {
		req := &contracts.GetMoviesRequest{}
		res, err := c.GetMovies(req)
		require.NoError(t, err)

		require.Equal(t, 3, res.Total)
		require.Equal(t, 1, res.Page)
		require.Equal(t, testPaginationSize, res.Size)
		require.Equal(t, []*contracts.Movie{&starWars.Movie, &harryPotter.Movie}, res.Items)

		req.Page = res.Page + 1
		res, err = c.GetMovies(req)
		require.NoError(t, err)

		require.Equal(t, 3, res.Total)
		require.Equal(t, 2, res.Page)
		require.Equal(t, testPaginationSize, res.Size)
		require.Equal(t, []*contracts.Movie{&lordOfTheRing.Movie}, res.Items)
	})

	t.Run("movies.DeleteMovie: success", func(t *testing.T) {
		req := &contracts.DeleteMovieRequest{
			ID: starWars.ID,
		}
		err := c.DeleteMovie(contracts.NewAuthenticated(req, johnDoeToken))
		require.NoError(t, err)

		starWars = getMovie(t, c, starWars.ID)
		require.Nil(t, starWars)
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
