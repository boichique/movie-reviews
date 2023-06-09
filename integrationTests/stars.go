package integrationTests

import (
	"net/http"
	"testing"
	"time"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/contracts"
	"github.com/stretchr/testify/require"
)

func starsAPIChecks(t *testing.T, c *client.Client) {
	var lucas, hamill, mcgregor *contracts.Star
	t.Run("stars.CreateStar: success", func(t *testing.T) {
		cases := []struct {
			req  *contracts.CreateStarRequest
			addr **contracts.Star
		}{
			{
				req: &contracts.CreateStarRequest{
					FirstName:  "George",
					MiddleName: ptr("Walton"),
					LastName:   "Lucas",
					BirthDate:  time.Date(1944, time.May, 14, 0, 0, 0, 0, time.UTC),
					BirthPlace: ptr("Modesto, California6 U.S."),
					Bio:        ptr("Famous creator of Star Wars"),
				},
				addr: &lucas,
			},
			{
				req: &contracts.CreateStarRequest{
					FirstName:  "Mark",
					MiddleName: ptr("Richard"),
					LastName:   "Hamill",
					BirthDate:  time.Date(1951, time.September, 25, 0, 0, 0, 0, time.UTC),
					BirthPlace: ptr("Oakland, California6 U.S."),
				},
				addr: &hamill,
			},
			{
				req: &contracts.CreateStarRequest{
					FirstName:  "Ewan",
					MiddleName: ptr("Gordon"),
					LastName:   "McGregor",
					BirthDate:  time.Date(1971, time.March, 31, 0, 0, 0, 0, time.UTC),
					BirthPlace: ptr("Perth, Scotland"),
				},
				addr: &mcgregor,
			},
		}

		for _, cc := range cases {

			star, err := c.CreateStar(contracts.NewAuthenticated(cc.req, johnDoeToken))
			require.NoError(t, err)

			*cc.addr = star
			require.NotEmpty(t, star.ID)
			require.NotEmpty(t, star.CreatedAd)
		}
	})

	t.Run("stars.GetStar: success", func(t *testing.T) {
		for _, star := range []*contracts.Star{lucas, hamill, mcgregor} {
			s, err := c.GetStarByID(star.ID)
			require.NoError(t, err)
			require.Equal(t, star, s)
		}
	})

	t.Run("stars.GetStar: not found", func(t *testing.T) {
		nonExistingID := 1000
		_, err := c.GetStarByID(nonExistingID)
		requireNotFoundError(t, err, "star", "id", nonExistingID)
	})

	t.Run("stars.UpdateStar: success", func(t *testing.T) {
		req := &contracts.UpdateStarRequest{
			ID:         lucas.ID,
			FirstName:  "George",
			MiddleName: ptr("Walton"),
			LastName:   "Lucas",
			BirthDate:  time.Date(1944, time.May, 14, 0, 0, 0, 0, time.UTC),
			BirthPlace: ptr("Modesto, California U.S."),
			Bio:        ptr("Famous creator of Star Wars and other films"),
		}
		err := c.UpdateStar(contracts.NewAuthenticated(req, johnDoeToken))
		require.NoError(t, err)

		lucas = getStar(t, c, lucas.ID)
		require.Equal(t, req.Bio, lucas.Bio)
	})

	t.Run("stars.GetStars: success", func(t *testing.T) {
		req := &contracts.GetStarsPaginatedRequest{}
		res, err := c.GetStars(req)
		require.NoError(t, err)

		require.Equal(t, 3, res.Total)
		require.Equal(t, 1, res.Page)
		require.Equal(t, testPaginationSize, res.Size)
		require.Equal(t, []*contracts.Star{lucas, hamill}, res.Items)

		req.Page = res.Page + 1
		res, err = c.GetStars(req)
		require.NoError(t, err)

		require.Equal(t, 3, res.Total)
		require.Equal(t, 2, res.Page)
		require.Equal(t, testPaginationSize, res.Size)
		require.Equal(t, []*contracts.Star{mcgregor}, res.Items)
	})

	t.Run("stars.DeleteStar: success", func(t *testing.T) {
		req := &contracts.GetOrDeleteStarRequest{
			ID: lucas.ID,
		}
		err := c.DeleteStar(contracts.NewAuthenticated(req, johnDoeToken))
		require.NoError(t, err)

		lucas = getStar(t, c, lucas.ID)
		require.Nil(t, lucas)
	})
}

func getStar(t *testing.T, c *client.Client, id int) *contracts.Star {
	u, err := c.GetStarByID(id)
	if err != nil {
		cerr, ok := err.(*client.Error)
		require.True(t, ok)
		require.Equal(t, http.StatusNotFound, cerr.Code)
		return nil
	}

	return u
}
