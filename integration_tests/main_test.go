package integration_tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB     *sqlx.DB
	testDBConn string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       "casino_transactions",
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "123456",
			},
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready to accept connections"),
				wait.ForListeningPort("5432/tcp"),
			),
		},
		Started: true,
	})
	if err != nil {
		fmt.Printf("Failed to start postgres container: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			fmt.Printf("Failed to terminate postgres container: %v\n", err)
		}
	}()

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		fmt.Printf("Failed to get container host: %v\n", err)
		os.Exit(1)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Printf("Failed to get container port: %v\n", err)
		os.Exit(1)
	}

	testDBConn = fmt.Sprintf("host=%s port=%s user=postgres password=123456 dbname=casino_transactions sslmode=disable",
		host, port.Port())

	testDB, err = sqlx.Connect("postgres", testDBConn)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	if err := runMigrations(testDB); err != nil {
		fmt.Printf("Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func runMigrations(db *sqlx.DB) error {
	migrationsPath, err := getMigrationsPath()
	if err != nil {
		return err
	}

	sqlDB := db.DB

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func getMigrationsPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	var migrationsDir string
	if filepath.Base(wd) == "integration_tests" {
		migrationsDir = filepath.Join(wd, "..", "migrations", "postgresql")
	} else {
		migrationsDir = filepath.Join(wd, "migrations", "postgresql")
	}

	migrationsPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	return migrationsPath, nil
}

func GetTestDB() *sqlx.DB {
	return testDB
}

func CleanupDB(t *testing.T) {
	if testDB == nil {
		t.Fatal("testDB is not initialized")
	}
	_, err := testDB.Exec("TRUNCATE TABLE transaction_events")
	if err != nil {
		t.Fatalf("Failed to cleanup database: %v", err)
	}
}
