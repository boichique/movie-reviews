package client

import "github.com/boichique/movie-reviews/contracts"

func (c *Client) CreateGenre(req *contracts.AuthenticatedRequest[*contracts.CreateGenreRequest]) (*contracts.Genre, error) {
	var genre *contracts.Genre

	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetBody(req.Request).
		SetResult(&genre).
		Post(c.path("/api/genres"))

	return genre, err
}

func (c *Client) GetGenreByID(genreID int) (*contracts.Genre, error) {
	var genre contracts.Genre

	_, err := c.client.R().
		SetResult(&genre).
		Get(c.path("/api/genres/%d", genreID))

	return &genre, err
}

func (c *Client) GetGenres() ([]*contracts.Genre, error) {
	var genres []*contracts.Genre

	_, err := c.client.R().
		SetResult(&genres).
		Get(c.path("/api/genres"))

	return genres, err
}

func (c *Client) UpdateGenre(req *contracts.AuthenticatedRequest[*contracts.UpdateGenreRequest]) error {
	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetBody(req.Request).
		Put(c.path("/api/genres/%d", req.Request.GenreID))

	return err
}

func (c *Client) DeleteGenre(req *contracts.AuthenticatedRequest[*contracts.GetOrDeleteGenreRequest]) error {
	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetBody(req.Request).
		Delete(c.path("/api/genres/%d", req.Request.GenreID))

	return err
}
