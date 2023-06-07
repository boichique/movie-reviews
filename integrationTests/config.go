package integrationTests

import (
	"time"

	"github.com/boichique/movie-reviews/internal/config"
)

const testPaginationSize = 2

func getConfig(pgConnString string) *config.Config {
	return &config.Config{
		DBUrl: pgConnString,
		Port:  0,
		Jwt: config.JwtConfig{
			Secret:           "secret",
			AccessExpiration: time.Minute * 15,
		},
		Admin: config.AdminConfig{
			AdminName:     "admin",
			AdminPassword: "&dm1Npa$$",
			AdminEmail:    "admin@email.com",
		},
		Pagination: config.PaginationConfig{
			DefaultSize: testPaginationSize,
			MaxSize:     50,
		},
		Local:    true,
		LogLevel: "error",
	}
}
