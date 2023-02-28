package migrate

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

func NewPostgresMigrateInstance(pgUrl, migrationsDir string) (*migrate.Migrate, error) {
	dbInstance, err := sql.Open("postgres", pgUrl)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("sql.Open('postgres', %s)", pgUrl))
	}

	driver, err := postgres.WithInstance(dbInstance, &postgres.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "postgres.WithInstance(dbInstance, &postgres.Config{})")
	}

	migrationsPathUrl := fmt.Sprintf("file://%s", migrationsDir)
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPathUrl,
		"postgres",
		driver,
	)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("migrate.NewWithDatabaseInstance(%s, 'postgres', driver)", migrationsPathUrl))
	}

	return m, nil
}
