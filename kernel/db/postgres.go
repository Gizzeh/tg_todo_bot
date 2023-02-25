package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx"
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

func (d *PG) OpenPool() (*pgx.ConnPool, error) {
	poolConf := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     d.host,
			Port:     uint16(d.port),
			Database: d.database,
			User:     d.user,
			Password: d.password,
		},
		MaxConnections: d.maxConnections,
	}
	pgConn, err := pgx.NewConnPool(poolConf)
	if err != nil {
		return nil, errors.Wrap(err, "pgx.NewConnPool(poolConf)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err = pgConn.ExecEx(ctx, ";", &pgx.QueryExOptions{})
	if err != nil {
		return nil, errors.Wrap(err, `pgConn.ExecEx(ctx, ";", &pgx.QueryExOptions{})`)
	}

	return pgConn, nil
}
