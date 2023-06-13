package client

import "github.com/boichique/movie-reviews/contracts"

func (c *Client) GetMovie(movieID int) (*contracts.MovieDetails, error) {
	var m contracts.MovieDetails

	_, err := c.client.R().
		SetResult(&m).
		Get(c.path("/api/movies/%d", movieID))

	return &m, err
}

func (c *Client) GetMovies(req *contracts.GetMoviesRequest) (*contracts.PaginatedResponse[contracts.Movie], error) {
	var res contracts.PaginatedResponse[contracts.Movie]

	_, err := c.client.R().
		SetResult(&res).
		SetQueryParams(req.PaginatedRequest.ToQueryParams()).
		Get(c.path("/api/movies"))

	return &res, err
}

func (c *Client) CreateMovie(req *contracts.AuthenticatedRequest[*contracts.CreateMovieRequest]) (*contracts.MovieDetails, error) {
	var g *contracts.MovieDetails
	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetBody(req.Request).
		SetResult(&g).
		Post(c.path("/api/movies"))

	return g, err
}

func (c *Client) UpdateMovie(req *contracts.AuthenticatedRequest[*contracts.UpdateMovieRequest]) error {
	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(req.Request).
		Put(c.path("/api/movies/%d", req.Request.ID))

	return err
}

func (c *Client) DeleteMovie(req *contracts.AuthenticatedRequest[*contracts.DeleteMovieRequest]) error {
	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(req.Request).
		Delete(c.path("/api/movies/%d", req.Request.ID))

	return err
}
