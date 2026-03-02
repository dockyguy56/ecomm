package main

import (
	"log"

	"github.com/dockyguy56/ecomm/internal/adapters/postgresql/env"
	db "github.com/dockyguy56/ecomm/internal/adapters/postgresql/sqlx"
)

func main() {
	// Load environment variables
	dbString := env.GetString("GOOSE_DBSTRING", "user=postgres password=postgres host=localhost port=5432 dbname=ecomm sslmode=disable")

	database, err := db.NewDatabase(dbString)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	log.Println("successfully connected to database")
}
