package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/boichique/movie-reviews/client"
	"github.com/boichique/movie-reviews/internal/config"
	"github.com/boichique/movie-reviews/internal/server"
	"github.com/hashicorp/consul/sdk/testutil/retry"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	prepareInfrastructure(t, runServer)
}

func runServer(t *testing.T, pgConnString string) {
	cfg := getConfig(pgConnString)

	srv, err := server.New(context.Background(), cfg)
	require.NoError(t, err)
	defer srv.Close()

	go func() {
		if serr := srv.Start(); serr != http.ErrServerClosed {
			require.NoError(t, serr)
		}
	}()

	var port int
	retry.Run(t, func(r *retry.R) {
		port, err = srv.Port()
		if err != nil {
			require.NoError(r, err)
		}
	})

	tests(t, port, cfg)

	err = srv.Shutdown(context.Background())
	require.NoError(t, err)
}

func tests(t *testing.T, port int, cfg *config.Config) {
	addr := fmt.Sprintf("http://localhost:%d", port)
	c := client.New(addr)

	// template for test names:
	// [module].[client_method]: [expected result or condition]
	// For example:
	// auth.Login: wrong password
	// users.GetUsers: success
	authAPIChecks(t, c, cfg)
	usersAPIChecks(t, c, cfg)
	genresAPIChecks(t, c)
	starsAPIChecks(t, c)
	moviesAPIChecks(t, c)
}
