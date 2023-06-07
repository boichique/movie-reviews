package client

import "github.com/boichique/movie-reviews/contracts"

func (c *Client) RegisterUser(req *contracts.RegisterUserRequest) (*contracts.User, error) {
	var user contracts.User

	_, err := c.client.R().
		SetBody(req).
		SetResult(&user).
		Post(c.path("/api/auth/register"))

	return &user, err
}

func (c *Client) LoginUser(req *contracts.LoginUserRequest) (*contracts.LoginUserResponse, error) {
	var resp contracts.LoginUserResponse

	_, err := c.client.R().
		SetBody(req).
		SetResult(&resp).
		Post(c.path("/api/auth/login"))

	return &resp, err
}
