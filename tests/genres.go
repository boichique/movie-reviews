package tests

import (
	"net/http"
	"testing"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/contracts"
	"github.com/stretchr/testify/require"
)

var (
	Action    *contracts.Genre
	Adventure *contracts.Genre
	SciFi     *contracts.Genre
	Drama     *contracts.Genre
	Spooky    *contracts.Genre
)

func genresAPIChecks(t *testing.T, c *client.Client) {
	t.Run("genres.GetGenres: empty", func(t *testing.T) {
		genres, err := c.GetGenres()
		require.NoError(t, err)
		require.Empty(t, genres)
	})

	t.Run("genres.CreateGenre: success: Action by Admin, Drama, Spooky, Adventure and SciFi by John Doe", func(t *testing.T) {
		cases := []struct {
			name  string
			token string
			addr  **contracts.Genre
		}{
			{"Action", adminToken, &Action},
			{"Drama", johnDoeToken, &Drama},
			{"Spooky", johnDoeToken, &Spooky},
			{"Adventure", johnDoeToken, &Adventure},
			{"SciFi", johnDoeToken, &SciFi},
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
			Name: Action.Name,
		}
		_, err := c.CreateGenre(contracts.NewAuthenticated(req, johnDoeToken))
		requireAlreadyExistsError(t, err, "genre", "name", Action.Name)
	})

	t.Run("genres.GetGenres: five genres", func(t *testing.T) {
		genres, err := c.GetGenres()
		require.NoError(t, err)
		require.Equal(t, []*contracts.Genre{Action, Drama, Spooky, Adventure, SciFi}, genres)
	})

	t.Run("genres.GetGenre: success", func(t *testing.T) {
		g, err := c.GetGenreByID(Spooky.ID)
		require.NoError(t, err)
		require.Equal(t, Spooky, g)
	})

	t.Run("genres.GetGenre: not found", func(t *testing.T) {
		nonExistingID := 1000
		_, err := c.GetGenreByID(nonExistingID)
		requireNotFoundError(t, err, "genre", "id", nonExistingID)
	})

	t.Run("genres.UpdateGenre: success", func(t *testing.T) {
		req := &contracts.UpdateGenreRequest{
			GenreID: Spooky.ID,
			Name:    "Horror",
		}
		err := c.UpdateGenre(contracts.NewAuthenticated(req, johnDoeToken))
		require.NoError(t, err)
		Spooky = getGenre(t, c, Spooky.ID)
		require.Equal(t, req.Name, Spooky.Name)
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
			GenreID: Spooky.ID,
		}
		err := c.DeleteGenre(contracts.NewAuthenticated(req, johnDoeToken))
		require.NoError(t, err)
		Spooky = getGenre(t, c, Spooky.ID)
		require.Nil(t, Spooky)
	})

	t.Run("genres.GetGenres: four genres", func(t *testing.T) {
		genres, err := c.GetGenres()
		require.NoError(t, err)
		require.Equal(t, []*contracts.Genre{Action, Drama, Adventure, SciFi}, genres)
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
