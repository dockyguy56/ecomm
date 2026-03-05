package db

import (
	"context"

	// "github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // needed for postgres driver "go get github.com/lib/pq@latest"
)

type Database struct {
	db *sqlx.DB
}

func NewDatabase(ctx context.Context, dbString string) (*Database, error) {
	db, err := sqlx.Connect("postgres", dbString)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) GetDB() *sqlx.DB {
	return d.db
}

func (d *Database) Close() error {
	return d.db.Close()
}
