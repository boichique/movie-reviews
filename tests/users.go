package tests

import (
	"net/http"
	"testing"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/contracts"
	"github.com/boichique/movie-reviews/internal/config"
	"github.com/boichique/movie-reviews/internal/modules/users"
	"github.com/stretchr/testify/require"
)

func usersAPIChecks(t *testing.T, c *client.Client, cfg *config.Config) {
	t.Run("users.GetUserByUsername: admin", func(t *testing.T) {
		u, err := c.GetUserByUsername(cfg.Admin.AdminName)
		require.NoError(t, err)
		admin = u

		require.Equal(t, cfg.Admin.AdminName, u.Username)
		require.Equal(t, cfg.Admin.AdminEmail, u.Email)
		require.Equal(t, users.AdminRole, u.Role)
	})

	t.Run("users.GetUserByUsername: not found", func(t *testing.T) {
		_, err := c.GetUserByUsername("notfound")
		requireNotFoundError(t, err, "user", "username", "notfound")
	})

	t.Run("users.GetUserByID: admin", func(t *testing.T) {
		u, err := c.GetUserByID(admin.ID)
		require.NoError(t, err)

		require.Equal(t, admin, u)
	})

	t.Run("users.GetUserByID: not found", func(t *testing.T) {
		nonExistingID := 1000
		_, err := c.GetUserByID(nonExistingID)
		requireNotFoundError(t, err, "user", "id", nonExistingID)
	})

	t.Run("users.UpdateUserBio: success", func(t *testing.T) {
		bio := "I'm John Doe"
		req := &contracts.UpdateUserBioRequest{
			UserID: johnDoe.ID,
			Bio:    &bio,
		}
		err := c.UpdateUserBio(contracts.NewAuthenticated(req, johnDoeToken))
		require.NoError(t, err)

		johnDoe = getUser(t, c, johnDoe.ID)
		require.Equal(t, bio, *johnDoe.Bio)
	})

	t.Run("users.UpdateUserBio: non-authenticated", func(t *testing.T) {
		bio := "I'm John Doe"
		req := &contracts.UpdateUserBioRequest{
			UserID: johnDoe.ID,
			Bio:    &bio,
		}
		err := c.UpdateUserBio(contracts.NewAuthenticated(req, ""))
		requireUnauthorizedError(t, err, "invalid or missing token")
	})

	t.Run("users.UpdateUserBio: another user", func(t *testing.T) {
		bio := "I'm John Doe"
		req := &contracts.UpdateUserBioRequest{
			UserID: johnDoe.ID + 1,
			Bio:    &bio,
		}
		err := c.UpdateUserBio(contracts.NewAuthenticated(req, johnDoeToken))
		requireForbiddenError(t, err, "insufficient permissions")
	})

	t.Run("users.UpdateUserBio: by admin", func(t *testing.T) {
		bio := "Updated by admin"
		req := &contracts.UpdateUserBioRequest{
			UserID: johnDoe.ID,
			Bio:    &bio,
		}

		err := c.UpdateUserBio(contracts.NewAuthenticated(req, adminToken))
		require.NoError(t, err)

		johnDoe = getUser(t, c, johnDoe.ID)
		require.Equal(t, bio, *johnDoe.Bio)
	})

	t.Run("users.UpdateUserRole: John Doe to editor", func(t *testing.T) {
		req := &contracts.UpdateUserRoleRequest{
			UserID: johnDoe.ID,
			Role:   users.EditorRole,
		}
		err := c.UpdateUserRole(contracts.NewAuthenticated(req, adminToken))
		require.NoError(t, err)

		johnDoe = getUser(t, c, johnDoe.ID)
		require.Equal(t, users.EditorRole, johnDoe.Role)

		// Have to re-login to become an editor
		johnDoeToken = login(t, c, johnDoe.Email, johnDoePass)
	})

	t.Run("users.UpdateUserRole: bad role", func(t *testing.T) {
		req := &contracts.UpdateUserRoleRequest{
			UserID: johnDoe.ID,
			Role:   "superuser",
		}
		err := c.UpdateUserRole(contracts.NewAuthenticated(req, adminToken))
		requireBadRequestError(t, err, "Role")
	})

	randomUser := registerRandomUser(t, c)
	t.Run("users.DeleteUser: another user", func(t *testing.T) {
		req := &contracts.GetOrDeleteUserRequest{
			UserID: randomUser.ID,
		}
		err := c.DeleteUser(contracts.NewAuthenticated(req, johnDoeToken))
		requireForbiddenError(t, err, "insufficient permissions")

		randomUser = getUser(t, c, randomUser.ID)
		require.NotNil(t, randomUser)
	})

	t.Run("users.DeleteUser: by admin", func(t *testing.T) {
		req := &contracts.GetOrDeleteUserRequest{
			UserID: randomUser.ID,
		}
		err := c.DeleteUser(contracts.NewAuthenticated(req, adminToken))
		require.NoError(t, err)

		randomUser = getUser(t, c, randomUser.ID)
		require.Nil(t, randomUser)
	})
}

func getUser(t *testing.T, c *client.Client, id int) *contracts.User {
	u, err := c.GetUserByID(id)
	if err != nil {
		cerr, ok := err.(*client.Error)
		require.True(t, ok)
		require.Equal(t, http.StatusNotFound, cerr.Code)
		return nil
	}

	return u
}
