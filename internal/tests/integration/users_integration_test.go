package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cptallergy/sidequest-api/internal/api"
	db "github.com/cptallergy/sidequest-api/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	dbUser     = "user"
	dbPassword = "password"
	dbHost     = "localhost"
	dbPort     = "5432"
	dbName     = "tests"
	dbSslMode  = "disable"
)

func setupTestDatabase(t *testing.T, ctx context.Context) (*postgres.PostgresContainer, string) {
	t.Setenv("DB_USER", dbUser)
	t.Setenv("DB_PASSWORD", dbPassword)
	t.Setenv("DB_HOST", dbHost)
	t.Setenv("DB_PORT", dbPort)
	t.Setenv("DB_NAME", dbName)
	t.Setenv("DB_SSLMODE", dbSslMode)

	ctr, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	defer testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	dbURL, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	migrationDB, err := sql.Open("pgx", dbURL)
	require.NoError(t, err)

	// Run all migrations
	err = goose.Up(migrationDB, "../../db/migrations")
	require.NoError(t, err)

	err = migrationDB.Close()
	require.NoError(t, err)

	// Create snapshot to restore to
	err = ctr.Snapshot(ctx)
	require.NoError(t, err)

	return ctr, dbURL
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	ctr, dbURL := setupTestDatabase(t, ctx)

	t.Run("Test inserting a user", func(t *testing.T) {
		t.Cleanup(func() {
			// 3. In each test, reset the DB to its snapshot state.
			restoreErr := ctr.Restore(ctx)
			require.NoError(t, restoreErr)
		})

		connPool, err := pgxpool.New(context.Background(), dbURL)
		require.NoError(t, err)
		defer connPool.Close()

		store := db.NewStore(connPool)
		app := &api.Application{
			Config: api.Config{},
			Store:  store,
		}

		testServer := httptest.NewServer(app.Mount())

		user, err := store.CreateUser(ctx, db.CreateUserParams{
			Username: "testuser",
			Email:    "test@email.com",
		})
		require.NoError(t, err)
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/users/%s", testServer.URL, user.Username))
		require.NoError(t, err)

		require.EqualValues(t, 200, resp.StatusCode)
	})

	// TODO add more complete tests
	//testServer := runTestServer()
	//defer testServer.Close()
	//
	//resp, err := http.Get(fmt.Sprintf("%s/quests", testServer.URL))
	//
	//if err != nil {
	//	t.Fatalf("Expected no error, got %v", err)
	//}
	//
	//if resp.StatusCode != 200 {
	//	t.Errorf("expected 200 got: %v", resp.StatusCode)
	//}
}
