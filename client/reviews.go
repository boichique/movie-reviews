package client

import "github.com/boichique/movie-reviews/contracts"

func (c *Client) CreateReview(req *contracts.AuthenticatedRequest[*contracts.CreateReviewRequest]) (*contracts.Review, error) {
	var review contracts.Review

	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetBody(req.Request).
		SetResult(&review).
		Post(c.path("/api/users/%d/reviews", req.Request.UserID))

	return &review, err
}

func (c *Client) GetReview(id int) (*contracts.Review, error) {
	var review contracts.Review

	_, err := c.client.R().
		SetResult(&review).
		Get(c.path("/api/reviews/%d", id))

	return &review, err
}

func (c *Client) GetReviews(req *contracts.GetReviewsRequest) (*contracts.PaginatedResponse[contracts.Review], error) {
	var res contracts.PaginatedResponse[contracts.Review]

	_, err := c.client.R().
		SetResult(&res).
		SetQueryParams(req.ToQueryParams()).
		Get(c.path("/api/reviews"))

	return &res, err
}

func (c *Client) UpdateReview(req *contracts.AuthenticatedRequest[*contracts.UpdateReviewRequest]) error {
	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetBody(req.Request).
		Put(c.path("/api/users/%d/reviews/%d", req.Request.UserID, req.Request.ReviewID))

	return err
}

func (c *Client) DeleteReview(req *contracts.AuthenticatedRequest[*contracts.DeleteReviewRequest]) error {
	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		Delete(c.path("/api/users/%d/reviews/%d", req.Request.UserID, req.Request.ReviewID))

	return err
}
