// Package dbtest provides database containers for unit and integration testing
package dbtest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang-todo-app/internal/core/dbpool"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Required for postgres
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Required for file source
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.uber.org/zap"
)

// NewPostgresTest creates a new postgres test container
func NewPostgresTest(ctx context.Context, logger *zap.Logger, migrationDir string, cfg *dbpool.Config) (*PostgresTest, error) {
	pgt := &PostgresTest{
		logger: logger,
		cfg:    cfg,
	}

	err := pgt.connect(ctx, migrationDir, cfg)
	if err != nil {
		return nil, err
	}

	return pgt, nil
}

// PostgresTest starts a postgres container to provide a database for unit testing
type PostgresTest struct {
	db           *pgxpool.Pool
	pool         *dockertest.Pool
	resource     *dockertest.Resource
	migrationDir string
	logger       *zap.Logger
	cfg          *dbpool.Config
}

const (
	maxContainerRunTimeSecs uint = 120
)

// connect starts up a test container of the database and connects to it
func (pgt *PostgresTest) connect(ctx context.Context, migrationDir string, cfg *dbpool.Config) (err error) {
	pgt.migrationDir = migrationDir

	// Create a new Docker pool
	pgt.pool, err = dockertest.NewPool("")
	if err != nil {
		pgt.logger.Error("Could not construct pool", zap.Error(err))
		return fmt.Errorf("could not construct pool: %w", err)
	}

	// Ping the Docker daemon
	err = pgt.pool.Client.Ping()
	if err != nil {
		pgt.logger.Error("Could not connect to Docker", zap.Error(err))
		return fmt.Errorf("could not connect to Docker: %w", err)
	}

	// Pull an image, create a container, and run it
	pgt.resource, err = pgt.pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16-alpine",
		Env: []string{
			fmt.Sprintf("POSTGRES_PASSWORD=%s", cfg.Password),
			fmt.Sprintf("POSTGRES_USER=%s", cfg.User),
			fmt.Sprintf("POSTGRES_DB=%s", cfg.DatabaseName),
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		config.Tmpfs = map[string]string{"/var/lib/postgresql": "rw"}
	})
	if err != nil {
		pgt.logger.Error("Could not start resource", zap.Error(err))
		return fmt.Errorf("could not start resource: %w", err)
	}

	// Extract the host and port for the container
	hostAndPort := strings.Split(pgt.resource.GetHostPort("5432/tcp"), ":")
	cfg.Host = hostAndPort[0]
	cfg.Port = hostAndPort[1]

	pgt.logger.Debug("Connecting to database", zap.String("url", cfg.ConnectString()))

	pgt.resource.Expire(maxContainerRunTimeSecs)

	// Retry connecting to the database
	pgt.pool.MaxWait = time.Duration(maxContainerRunTimeSecs) * time.Second
	if err = pgt.pool.Retry(func() error {
		pgt.db, err = dbpool.New(ctx, pgt.logger, cfg)
		if err != nil {
			return err
		}
		return pgt.db.Ping(ctx)
	}); err != nil {
		pgt.logger.Error("Could not connect to Docker", zap.Error(err))
		return fmt.Errorf("could not connect to Docker: %w", err)
	}

	return nil
}

// DB returns a database pool
func (pgt *PostgresTest) DB() *pgxpool.Pool {
	return pgt.db
}

// TearDown shuts down the container
func (pgt *PostgresTest) TearDown() error {
	pgt.db.Close()

	if err := pgt.pool.Purge(pgt.resource); err != nil {
		return fmt.Errorf("could not purge resource: %w", err)
	}
	return nil
}

// MigrateUp applies the database migration to the latest version
func (pgt *PostgresTest) MigrateUp() error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", pgt.migrationDir),
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			pgt.cfg.User,
			pgt.cfg.Password,
			pgt.cfg.Host,
			pgt.cfg.Port,
			pgt.cfg.DatabaseName,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	pgt.logger.Info("Database migrations applied successfully")
	return nil
}

// MigrateReset rolls back all migrations
func (pgt *PostgresTest) MigrateReset() error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", pgt.migrationDir),
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			pgt.cfg.User,
			pgt.cfg.Password,
			pgt.cfg.Host,
			pgt.cfg.Port,
			pgt.cfg.DatabaseName,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	pgt.logger.Info("Database migrations rolled back successfully")
	return nil
}
