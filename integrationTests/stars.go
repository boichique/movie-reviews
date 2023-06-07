package integrationTests

import (
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
}
