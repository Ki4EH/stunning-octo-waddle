package db

import (
	"context"
	"fmt"
	"github.com/Ki4EH/stunning-octo-waddle/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

func NewPostgresDB(cfg *config.Config) (*pgxpool.Pool, error) {

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?pool_max_conns=20",
		cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName)

	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("%v Unable to create connection pool: %v\n", os.Stderr, err)
	}

	err = dbpool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%v Unable to ping database: %v\n", os.Stderr, err)
	}

	return dbpool, nil
}
