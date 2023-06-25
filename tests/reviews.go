package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/contracts"
	"github.com/stretchr/testify/require"
)

func reviewsAPIChecks(t *testing.T, c *client.Client) {
	reviewer1 := registerRandomUser(t, c)
	reviewer2 := registerRandomUser(t, c)
	reviewer1Token := login(t, c, reviewer1.Email, standardPassword)
	reviewer2Token := login(t, c, reviewer2.Email, standardPassword)

	var review1, review2, review3 *contracts.Review
	t.Run("reviews.CreateReview: success", func(t *testing.T) {
		cases := []struct {
			req   *contracts.CreateReviewRequest
			token string
			addr  **contracts.Review
		}{
			{
				req: &contracts.CreateReviewRequest{
					MovieID: StarWars.ID,
					UserID:  reviewer1.ID,
					Rating:  10,
					Title:   "Legendary piece of cinema",
					Content: "I love the original Star Wars films! They're a magical experience with great music and " +
						"sounds. They were made amazingly for their time. Some parts can be boring, but overall " +
						"they're glorious. I didn't understand the hype until a few years ago, but now I'm happy with " +
						"all the films, including the new ones.",
				},
				token: reviewer1Token,
				addr:  &review1,
			},
			{
				req: &contracts.CreateReviewRequest{
					MovieID: StarWars.ID,
					UserID:  reviewer2.ID,
					Rating:  9,
					Title:   "A long time ago in a decade without CGI...",
					Content: "A timeless classic with impressive practical effects, despite outdated CGI. A must-watch " +
						"for fans of the franchise and a testament to its enduring greatness.",
				},
				token: reviewer2Token,
				addr:  &review2,
			},
			{
				req: &contracts.CreateReviewRequest{
					MovieID: StarTrek.ID,
					UserID:  reviewer1.ID,
					Rating:  8,
					Title:   "The Emotion Picture...",
					Content: "I'll write the review later. Sorry",
				},
				token: reviewer1Token,
				addr:  &review3,
			},
		}

		for _, cc := range cases {
			review, err := c.CreateReview(contracts.NewAuthenticated(cc.req, cc.token))
			require.NoError(t, err)

			*cc.addr = review
		}
	})

	t.Run("reviews.CreateReview: already exists", func(t *testing.T) {
		req := &contracts.CreateReviewRequest{
			MovieID: review1.MovieID,
			UserID:  review1.UserID,
			Rating:  10,
			Title:   "Legendary movie",
			Content: "Just watch it. It's great.",
		}

		_, err := c.CreateReview(contracts.NewAuthenticated(req, reviewer1Token))
		requireAlreadyExistsError(t, err, "review", "(movie_id,user_id)", fmt.Sprintf("(%d,%d)", req.MovieID, req.UserID))
	})

	t.Run("reviews.GetReview: success", func(t *testing.T) {
		for _, review := range []*contracts.Review{review1, review2, review3} {
			r, err := c.GetReview(review.ID)
			require.NoError(t, err)

			require.Equal(t, review, r)
		}
	})

	t.Run("reviews.GetReviews: success", func(t *testing.T) {
		cases := []struct {
			req *contracts.GetReviewsRequest
			exp []*contracts.Review
		}{
			{
				req: &contracts.GetReviewsRequest{
					MovieID: ptr(StarWars.ID),
				},
				exp: []*contracts.Review{review1, review2},
			},
			{
				req: &contracts.GetReviewsRequest{
					MovieID: ptr(StarTrek.ID),
				},
				exp: []*contracts.Review{review3},
			},
			{
				req: &contracts.GetReviewsRequest{
					UserID: ptr(reviewer1.ID),
				},
				exp: []*contracts.Review{review1, review3},
			},
		}

		for _, cc := range cases {
			res, err := c.GetReviews(cc.req)
			require.NoError(t, err)

			require.Equal(t, cc.exp, res.Items)
		}
	})

	t.Run("reviews.GetReviews: no movieID or userID specified", func(t *testing.T) {
		_, err := c.GetReviews(&contracts.GetReviewsRequest{})
		requireBadRequestError(t, err, "either movie_id or user_id must be provided")
	})

	t.Run("reviews.UpdateReview: success", func(t *testing.T) {
		req := &contracts.UpdateReviewRequest{
			ReviewID: review3.ID,
			UserID:   review3.UserID,
			Rating:   review3.Rating,
			Title:    review3.Title,
			Content: "Boldly going where no man (or woman) has gone before, climb aboard the Enterprise and let " +
				"it fly and soar, as old friends gather, reunite, off to battle and to fight, strange new " +
				"worlds, civilisations to explore.",
		}

		err := c.UpdateReview(contracts.NewAuthenticated(req, reviewer1Token))
		require.NoError(t, err)

		review3 = getReview(t, c, review3.ID)
		require.Equal(t, req.Content, review3.Content)
	})

	t.Run("reviews.DeleteReview: not found", func(t *testing.T) {
		nonExistingID := 10000
		req := &contracts.DeleteReviewRequest{
			ReviewID: nonExistingID,
			UserID:   review3.UserID,
		}
		err := c.DeleteReview(contracts.NewAuthenticated(req, reviewer1Token))
		requireNotFoundError(t, err, "review", "id", nonExistingID)
	})

	t.Run("reviews.DeleteReview: mismatch between token and path", func(t *testing.T) {
		req := &contracts.DeleteReviewRequest{
			ReviewID: review3.ID,
			UserID:   reviewer1.ID,
		}
		err := c.DeleteReview(contracts.NewAuthenticated(req, reviewer2Token))
		requireForbiddenError(t, err, "insufficient permissions")
	})

	t.Run("reviews.DeleteReview: owned by another user", func(t *testing.T) {
		req := &contracts.DeleteReviewRequest{
			ReviewID: review3.ID,
			UserID:   reviewer2.ID,
		}
		err := c.DeleteReview(contracts.NewAuthenticated(req, reviewer2Token))
		requireForbiddenError(t, err, "review with id 3 is not owned by user with id 7")
	})

	t.Run("reviews.DeleteReview: success", func(t *testing.T) {
		req := &contracts.DeleteReviewRequest{
			ReviewID: review3.ID,
			UserID:   reviewer1.ID,
		}
		err := c.DeleteReview(contracts.NewAuthenticated(req, reviewer1Token))
		require.NoError(t, err)

		review3 = getReview(t, c, review3.ID)
		require.Nil(t, review3)
	})
}

func getReview(t *testing.T, c *client.Client, reviewID int) *contracts.Review {
	review, err := c.GetReview(reviewID)
	if err != nil {
		cerr, ok := err.(*client.Error)
		require.True(t, ok)
		require.Equal(t, http.StatusNotFound, cerr.Code)
		return nil
	}

	return review
}
