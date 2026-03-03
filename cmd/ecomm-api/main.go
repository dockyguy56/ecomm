package main

import (
	"log"

	db "github.com/dockyguy56/ecomm/internal/adapters/postgresql/sqlx"
	"github.com/dockyguy56/ecomm/internal/env"
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
