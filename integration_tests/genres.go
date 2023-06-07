package integration_tests

import (
	"net/http"
	"testing"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/contracts"
	"github.com/stretchr/testify/require"
)

func genresApiChecks(t *testing.T, c *client.Client) {
	t.Run("genres.GetGenres: empty", func(t *testing.T) {
		genres, err := c.GetGenres()
		require.NoError(t, err)
		require.Empty(t, genres)
	})

	var action, drama, spooky *contracts.Genre
	t.Run("genres.CreateGenre: success: Action by Admin, Drama and Spooky by John Doe", func(t *testing.T) {
		cases := []struct {
			name  string
			token string
			addr  **contracts.Genre
		}{
			{"Action", adminToken, &action},
			{"Drama", johnDoeToken, &drama},
			{"Spooky", johnDoeToken, &spooky},
		}

		for _, cc := range cases {
			req := &contracts.CreateGenreRequest{
				Name: cc.name,
			}
			g, err := c.CreateGenre(contracts.NewAuthenticated(req, cc.token))
			require.NoError(t, err)

			*cc.addr = g
			require.NotEmpty(t, g.ID)
			require.Equal(t, req.Name, g.Name)
		}
	})

	t.Run("genres.CreateGenre: short name", func(t *testing.T) {
		req := &contracts.CreateGenreRequest{
			Name: "Oh",
		}
		_, err := c.CreateGenre(contracts.NewAuthenticated(req, johnDoeToken))
		requireBadRequestError(t, err, "Name")
	})

	t.Run("genres.CreateGenre: existing name", func(t *testing.T) {
		req := &contracts.CreateGenreRequest{
			Name: action.Name,
		}
		_, err := c.CreateGenre(contracts.NewAuthenticated(req, johnDoeToken))
		requireAlreadyExistsError(t, err, "genre", "name", action.Name)
	})

	t.Run("genres.GetGenres: three genres", func(t *testing.T) {
		genres, err := c.GetGenres()
		require.NoError(t, err)
		require.Equal(t, []*contracts.Genre{action, drama, spooky}, genres)
	})

	t.Run("genres.GetGenreByID: success", func(t *testing.T) {
		g, err := c.GetGenreByID(spooky.ID)
		require.NoError(t, err)

		require.Equal(t, spooky, g)
	})

	t.Run("genres.GetGenreByID: not found", func(t *testing.T) {
		nonExistingId := 1000
		_, err := c.GetGenreByID(nonExistingId)
		requireNotFoundError(t, err, "genre", "id", nonExistingId)
	})

	t.Run("genres.UpdateGenre: success", func(t *testing.T) {
		req := &contracts.UpdateGenreRequest{
			GenreID: spooky.ID,
			Name:    "Horror",
		}
		err := c.UpdateGenre(contracts.NewAuthenticated(req, johnDoeToken))
		require.NoError(t, err)

		spooky = getGenre(t, c, spooky.ID)
		require.Equal(t, req.Name, spooky.Name)
	})

	t.Run("genres.UpdateGenre: not found", func(t *testing.T) {
		nonExistingID := 1000
		req := &contracts.UpdateGenreRequest{
			GenreID: nonExistingID,
			Name:    "Horror",
		}
		err := c.UpdateGenre(contracts.NewAuthenticated(req, johnDoeToken))
		requireNotFoundError(t, err, "genre", "id", nonExistingID)
	})

	t.Run("genres.DeleteGenre: success", func(t *testing.T) {
		req := &contracts.GetOrDeleteGenreRequest{
			GenreID: spooky.ID,
		}
		err := c.DeleteGenre(contracts.NewAuthenticated(req, johnDoeToken))
		require.NoError(t, err)

		spooky = getGenre(t, c, spooky.ID)
		require.Nil(t, spooky)
	})

	t.Run("genres.GetGenres: two genres", func(t *testing.T) {
		genres, err := c.GetGenres()
		require.NoError(t, err)
		require.Equal(t, []*contracts.Genre{action, drama}, genres)
	})
}

func getGenre(t *testing.T, c *client.Client, id int) *contracts.Genre {
	u, err := c.GetGenreByID(id)
	if err != nil {
		cerr, ok := err.(*client.Error)
		require.True(t, ok)
		require.Equal(t, http.StatusNotFound, cerr.Code)
		return nil
	}

	return u
}
