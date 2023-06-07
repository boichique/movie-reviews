package client

import "github.com/boichique/movie-reviews/contracts"

func (c *Client) CreateStar(req *contracts.AuthenticatedRequest[*contracts.CreateStarRequest]) (*contracts.Star, error) {
	var star *contracts.Star

	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetBody(req.Request).
		SetResult(&star).
		Post(c.path("/api/stars"))

	return star, err
}

func (c *Client) GetStarByID(starID int) (*contracts.Star, error) {
	var star contracts.Star

	_, err := c.client.R().
		SetResult(&star).
		Get(c.path("/api/stars/%d", starID))

	return &star, err
}

func (c *Client) GetStars() ([]*contracts.Star, error) {
	var stars []*contracts.Star

	_, err := c.client.R().
		SetResult(&stars).
		Get(c.path("/api/stars"))

	return stars, err
}

func (c *Client) DeleteStar(req *contracts.AuthenticatedRequest[*contracts.GetOrDeleteStarRequest]) error {
	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetBody(req.Request).
		Delete(c.path("api/stars/%d", req.Request.ID))

	return err
}
