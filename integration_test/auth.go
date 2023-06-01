package integration_tests

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/contracts"
	"github.com/boichique/movie-reviews/internal/config"
	"github.com/boichique/movie-reviews/internal/modules/users"
	"github.com/stretchr/testify/require"
)

const standardPassword = "secuR3P@ss"

var (
	johnDoe      *contracts.User // eventually going to be an editor
	johnDoePass  = standardPassword
	johnDoeToken string

	admin      *contracts.User
	adminToken string
)

func authApiChecks(t *testing.T, c *client.Client, cfg *config.Config) {
	t.Run("auth.Register: success", func(t *testing.T) {
		req := &contracts.RegisterUserRequest{
			Username: "johndoe",
			Email:    "johndoe@example.com",
			Password: johnDoePass,
		}
		u, err := c.RegisterUser(req)
		require.NoError(t, err)
		johnDoe = u

		require.Equal(t, req.Username, u.Username)
		require.Equal(t, req.Email, u.Email)
		require.Equal(t, users.UserRole, u.Role)
	})

	t.Run("auth.Register: short username", func(t *testing.T) {
		req := &contracts.RegisterUserRequest{
			Username: "joh",
			Email:    "joh@example.com",
			Password: standardPassword,
		}
		_, err := c.RegisterUser(req)
		requireBadRequestError(t, err, "Username")
	})

	t.Run("auth.Register: existing username", func(t *testing.T) {
		req := &contracts.RegisterUserRequest{
			Username: johnDoe.Username,
			Email:    "johndoe_another@example.com",
			Password: standardPassword,
		}
		_, err := c.RegisterUser(req)
		requireAlreadyExistsError(t, err, "user", "username", johnDoe.Username)
	})

	t.Run("auth.Register: existing email", func(t *testing.T) {
		req := &contracts.RegisterUserRequest{
			Username: "another_john",
			Email:    johnDoe.Email,
			Password: standardPassword,
		}
		_, err := c.RegisterUser(req)
		requireAlreadyExistsError(t, err, "user", "email", johnDoe.Email)
	})

	t.Run("auth.Login: success: John Doe", func(t *testing.T) {
		req := &contracts.LoginUserRequest{
			Email:    johnDoe.Email,
			Password: johnDoePass,
		}
		res, err := c.LoginUser(req)
		require.NoError(t, err)
		require.NotEmpty(t, res.AccessToken)
		johnDoeToken = res.AccessToken
	})

	t.Run("auth.Login: success: admin", func(t *testing.T) {
		req := &contracts.LoginUserRequest{
			Email:    cfg.Admin.AdminEmail,
			Password: cfg.Admin.AdminPassword,
		}
		res, err := c.LoginUser(req)
		require.NoError(t, err)
		require.NotEmpty(t, res.AccessToken)
		adminToken = res.AccessToken
	})

	t.Run("auth.Login: wrong password", func(t *testing.T) {
		req := &contracts.LoginUserRequest{
			Email:    johnDoe.Email,
			Password: johnDoePass + "wrong",
		}
		_, err := c.LoginUser(req)
		requireUnauthorizedError(t, err, "invalid password")
	})

	t.Run("auth.Login: wrong email", func(t *testing.T) {
		req := &contracts.LoginUserRequest{
			Email:    "nonexisting@mail.com",
			Password: standardPassword,
		}

		_, err := c.LoginUser(req)
		requireNotFoundError(t, err, "user", "email", req.Email)
	})
}

func registerRandomUser(t *testing.T, c *client.Client) *contracts.User {
	r := rand.Intn(10000)

	req := &contracts.RegisterUserRequest{
		Username: fmt.Sprintf("random_%d", r),
		Email:    fmt.Sprintf("random_%d@mail.com", r),
		Password: standardPassword,
	}
	u, err := c.RegisterUser(req)
	require.NoError(t, err)

	return u
}

func login(t *testing.T, c *client.Client, email, password string) string {
	req := &contracts.LoginUserRequest{
		Email:    email,
		Password: password,
	}
	res, err := c.LoginUser(req)
	require.NoError(t, err)

	return res.AccessToken
}
