package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // needed for postgres driver "go get github.com/lib/pq@latest"
)

type Database struct {
	db *sqlx.DB
}

func NewDatabase(dbString string) (*Database, error) {
	db, err := sqlx.Connect("postgres", dbString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return &Database{db: db}, nil
}

func (d *Database) GetDB() *sqlx.DB {
	return d.db
}

func (d *Database) Close() error {
	return d.db.Close()
}
