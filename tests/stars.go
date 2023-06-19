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
	MarkHamill     *contracts.StarDetails
	EvanMcGregor   *contracts.StarDetails
	GeorgeLucas    *contracts.StarDetails
	WilliamShatner *contracts.StarDetails
)

func starsAPIChecks(t *testing.T, c *client.Client) {
	t.Run("stars.CreateStar: success", func(t *testing.T) {
		cases := []struct {
			req  *contracts.CreateStarRequest
			addr **contracts.StarDetails
		}{
			{
				req: &contracts.CreateStarRequest{
					FirstName:  "George",
					MiddleName: ptr("Walton"),
					LastName:   "Lucas",
					BirthDate:  time.Date(1944, time.May, 14, 0, 0, 0, 0, time.UTC),
					BirthPlace: ptr("Modesto, California, U.S."),
					Bio:        ptr("Famous creator of Star Wars"),
				},
				addr: &GeorgeLucas,
			},
			{
				req: &contracts.CreateStarRequest{
					FirstName:  "Mark",
					MiddleName: ptr("Richard"),
					LastName:   "Hamill",
					BirthDate:  time.Date(1951, time.September, 25, 0, 0, 0, 0, time.UTC),
					BirthPlace: ptr("Oakland, California, U.S."),
				},
				addr: &MarkHamill,
			},
			{
				req: &contracts.CreateStarRequest{
					FirstName:  "Ewan",
					MiddleName: ptr("Gordon"),
					LastName:   "McGregor",
					BirthDate:  time.Date(1971, time.March, 31, 0, 0, 0, 0, time.UTC),
					BirthPlace: ptr("Perth, Scotland"),
				},
				addr: &EvanMcGregor,
			},
			{
				req: &contracts.CreateStarRequest{
					FirstName:  "William",
					MiddleName: ptr("Alan"),
					LastName:   "Shatner",
					BirthDate:  time.Date(1931, time.March, 22, 0, 0, 0, 0, time.UTC),
					BirthPlace: ptr("Montreal, Quebec, Canada"),
				},
				addr: &WilliamShatner,
			},
		}

		for _, cc := range cases {
			star, err := c.CreateStar(contracts.NewAuthenticated(cc.req, johnDoeToken))
			require.NoError(t, err)

			*cc.addr = star
			require.NotEmpty(t, star.ID)
			require.NotEmpty(t, star.CreatedAt)
		}
	})

	t.Run("stars.GetStar: success", func(t *testing.T) {
		for _, star := range []*contracts.StarDetails{GeorgeLucas, MarkHamill, EvanMcGregor, WilliamShatner} {
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

	t.Run("stars.GetStars: success", func(t *testing.T) {
		req := &contracts.GetStarsPaginatedRequest{}
		res, err := c.GetStars(req)
		require.NoError(t, err)

		require.Equal(t, 4, res.Total)
		require.Equal(t, 1, res.Page)
		require.Equal(t, testPaginationSize, res.Size)
		require.Equal(t, []*contracts.Star{&GeorgeLucas.Star, &MarkHamill.Star}, res.Items)

		req.Page = res.Page + 1
		res, err = c.GetStars(req)
		require.NoError(t, err)

		require.Equal(t, 4, res.Total)
		require.Equal(t, 2, res.Page)
		require.Equal(t, testPaginationSize, res.Size)
		require.Equal(t, []*contracts.Star{&EvanMcGregor.Star, &WilliamShatner.Star}, res.Items)
	})

	t.Run("stars.UpdateStar: success", func(t *testing.T) {
		req := &contracts.UpdateStarRequest{
			StarID:     EvanMcGregor.ID,
			FirstName:  EvanMcGregor.FirstName,
			MiddleName: EvanMcGregor.MiddleName,
			LastName:   EvanMcGregor.LastName,
			BirthDate:  EvanMcGregor.BirthDate,
			BirthPlace: EvanMcGregor.BirthPlace,
			DeathDate:  EvanMcGregor.DeathDate,
			Bio:        ptr(`Acclaimed Scottish actor known for "Trainspotting," "Moulin Rouge!," and Obi-Wan Kenobi in "Star Wars."`),
		}

		err := c.UpdateStar(contracts.NewAuthenticated(req, johnDoeToken))
		require.NoError(t, err)

		EvanMcGregor = getStar(t, c, EvanMcGregor.ID)
		require.Equal(t, req.Bio, EvanMcGregor.Bio)
	})

	t.Run("stars.DeleteStar: success", func(t *testing.T) {
		star := createRandomStar(t, c, johnDoeToken)
		err := c.DeleteStar(contracts.NewAuthenticated(&contracts.DeleteStarRequest{StarID: star.ID}, johnDoeToken))
		require.NoError(t, err)

		star = getStar(t, c, star.ID)
		require.Nil(t, star)
	})
}

func getStar(t *testing.T, c *client.Client, id int) *contracts.StarDetails {
	u, err := c.GetStarByID(id)
	if err != nil {
		cerr, ok := err.(*client.Error)
		require.True(t, ok)
		require.Equal(t, http.StatusNotFound, cerr.Code)
		return nil
	}

	return u
}

func createRandomStar(t *testing.T, c *client.Client, token string) *contracts.StarDetails {
	r := rand.Intn(10000)

	star, err := c.CreateStar(contracts.NewAuthenticated(&contracts.CreateStarRequest{
		FirstName:  fmt.Sprintf("First Name %d", r),
		MiddleName: ptr(fmt.Sprintf("Middle Name %d", r)),
		LastName:   fmt.Sprintf("Last Name %d", r),
		BirthDate:  time.Date(1971, time.March, 31, 0, 0, 0, 0, time.UTC),
		BirthPlace: ptr(fmt.Sprintf("Birth Place %d", r)),
	}, token))
	require.NoError(t, err)

	return star
}
