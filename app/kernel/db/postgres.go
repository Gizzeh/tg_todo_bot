package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"time"
)

const pgMaxConnections = 5

type PG struct {
	host           string
	port           int
	database       string
	user           string
	password       string
	maxConnections int // const
}

func NewPG(host string, port int, dbName string, user string, password string) *PG {
	return &PG{
		host:           host,
		port:           port,
		database:       dbName,
		user:           user,
		password:       password,
		maxConnections: pgMaxConnections,
	}
}

func (d *PG) GetConnUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		d.user,
		d.password,
		d.host,
		d.port,
		d.database,
	)
}

func (d *PG) OpenPool() (*pgxpool.Pool, error) {
	connUrl := d.GetConnUrl()
	pgPool, err := pgxpool.Connect(context.Background(), connUrl)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(`pgxpool.New(connUrl = '%s')`, connUrl))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err = pgPool.Exec(ctx, ";")
	if err != nil {
		return nil, errors.Wrap(err, `pgPool.Exec(ctx, ";")`)
	}

	return pgPool, nil
}
