package client

import "github.com/boichique/movie-reviews/contracts"

func (c *Client) GetUserByID(userID int) (*contracts.User, error) {
	var user contracts.User

	_, err := c.client.R().
		SetResult(&user).
		Get(c.path("/api/users/%d", userID))

	return &user, err
}

func (c *Client) GetUserByUsername(username string) (*contracts.User, error) {
	var user contracts.User

	_, err := c.client.R().
		SetResult(&user).
		Get(c.path("/api/users/username/%s", username))

	return &user, err
}

func (c *Client) UpdateUserBio(req *contracts.AuthenticatedRequest[*contracts.UpdateUserBioRequest]) error {
	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(req.Request).
		Put(c.path("/api/users/%d", req.Request.UserID))

	return err
}

func (c *Client) UpdateUserRole(req *contracts.AuthenticatedRequest[*contracts.UpdateUserRoleRequest]) error {
	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(req.Request).
		Put(c.path("/api/users/%d/role/%s", req.Request.UserID, req.Request.Role))

	return err
}

func (c *Client) DeleteUser(req *contracts.AuthenticatedRequest[*contracts.DeleteUserRequest]) error {
	_, err := c.client.R().
		SetAuthToken(req.AccessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(req.Request).
		Delete(c.path("/api/users/%d", req.Request.UserID))

	return err
}
