package e2e

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/111zxc/pr-review-service/internal/app"
	"github.com/111zxc/pr-review-service/internal/config"
	"github.com/111zxc/pr-review-service/internal/handler"
	pg "github.com/111zxc/pr-review-service/internal/repository/postgres"
	"github.com/111zxc/pr-review-service/internal/service"
)

type TestEnv struct {
	Ctx       context.Context
	DB        *pgxpool.Pool
	Server    *httptest.Server
	Container tc.Container
}

func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	ctx := context.Background()

	port := nat.Port("5432/tcp")

	req := tc.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{port.Port()},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForSQL(
			port,
			"pgx",
			func(host string, p nat.Port) string {
				return fmt.Sprintf(
					"postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
					host, p.Port(),
				)
			},
		).WithStartupTimeout(20 * time.Second),
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start postgres: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to host testcontainer: %v", err)
	}
	mapped, err := container.MappedPort(ctx, port)
	if err != nil {
		t.Fatalf("Failed to establish mapped port: %v", err)
	}

	dbURL := fmt.Sprintf(
		"postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
		host, mapped.Port(),
	)

	os.Setenv("DB_HOST", host)           //nolint:errcheck
	os.Setenv("DB_PORT", mapped.Port())  //nolint:errcheck
	os.Setenv("DB_USER", "testuser")     //nolint:errcheck
	os.Setenv("DB_PASSWORD", "testpass") //nolint:errcheck
	os.Setenv("DB_NAME", "testdb")       //nolint:errcheck
	os.Setenv("DB_SSLMODE", "disable")   //nolint:errcheck

	os.Setenv("ENV", "test") //nolint:errcheck

	cfg := config.Load()

	pool, err := pg.NewDB(cfg)
	if err != nil {
		t.Fatalf("failed to init pgxpool: %v", err)
	}

	runMigrations(t, dbURL)

	tx := pg.NewTxManager(pool)

	userRepo := pg.NewUserRepository(pool)
	teamRepo := pg.NewTeamRepository(pool, tx)
	prRepo := pg.NewPullRequestRepository(pool, tx)
	prStatusRepo := pg.NewPRStatusRepository(pool)
	statsRepo := pg.NewStatsRepository(pool)
	eventRepo := pg.NewEventsRepository(pool)

	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo)
	prService := service.NewPullRequestService(prRepo, userRepo, teamRepo, prStatusRepo, eventRepo)
	statsService := service.NewStatsService(statsRepo)

	h := handler.New(teamService, userService, prService, statsService)

	router := app.NewRouter(h)
	server := httptest.NewServer(router)

	return &TestEnv{
		Ctx:       ctx,
		DB:        pool,
		Server:    server,
		Container: container,
	}
}

func TearDown(env *TestEnv) {
	if env.Server != nil {
		env.Server.Close()
	}
	if env.DB != nil {
		env.DB.Close()
	}
	if env.Container != nil {
		if err := env.Container.Terminate(env.Ctx); err != nil {
			return
		}
	}
}

func runMigrations(t *testing.T, dbURL string) {
	t.Helper()

	cmd := exec.Command("goose", "-dir", "../../migrations", "postgres", dbURL, "up")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("goose failed: %v\n%s", err, out)
	}
}
