package integration_tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresContainerName = "movie-reviews-it-postgres"
	postgresUser          = "user"
	postgresPassword      = "pass"
	postgresDb            = "moviereviewsdb"
)

func prepareInfrastructure(t *testing.T, runFunc func(t *testing.T, connString string)) {
	// Start Postgres container
	postgres, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:  postgresContainerName,
			Image: "postgres:15-alpine",
			Env: map[string]string{
				"POSTGRES_USER":     postgresUser,
				"POSTGRES_PASSWORD": postgresPassword,
				"POSTGRES_DB":       postgresDb,
			},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	})
	require.NoError(t, err)
	defer cleanUp(t, postgres.Terminate)

	postgresPort, err := postgres.MappedPort(context.Background(), "5432")
	require.NoError(t, err)
	pgConnString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", postgresUser, postgresPassword, "localhost", postgresPort.Int(), postgresDb)

	// Run migrations
	time.Sleep(1 * time.Second) // It's a hack, but it works
	runMigrations(t, pgConnString)

	// Run tests
	runFunc(t, pgConnString)
}

func runMigrations(t *testing.T, connString string) {
	conn, err := pgx.Connect(context.Background(), connString)
	require.NoError(t, err)
	defer cleanUp(t, conn.Close)

	migrator, err := migrate.NewMigrator(context.Background(), conn, "schema_version")
	require.NoError(t, err)

	err = migrator.LoadMigrations(os.DirFS("../tern/migrations"))
	require.NoError(t, err)

	err = migrator.Migrate(context.Background())
	require.NoError(t, err)
}

func cleanUp(t *testing.T, cleanUpFunc func(context.Context) error) {
	require.NoError(t, cleanUpFunc(context.Background()))
}
