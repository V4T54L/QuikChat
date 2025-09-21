package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB(databaseURL string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v\n", err)
	}

	err = dbpool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to ping database: %v\n", err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")
	return dbpool, nil
}
