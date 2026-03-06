package main

import (
	"context"
	"log"

	db "github.com/dockyguy56/ecomm/internal/adapters/postgresql/sqlx"
	"github.com/dockyguy56/ecomm/internal/ecomm-api/handler"
	"github.com/dockyguy56/ecomm/internal/ecomm-api/server"
	"github.com/dockyguy56/ecomm/internal/ecomm-api/storer"
	"github.com/dockyguy56/ecomm/internal/env"
	"github.com/ianschenck/envflag"
)

const minSecretKeySize = 32

func main() {
	// Load environment variables
	secretKey := envflag.String("SECRET_KEY", "01234567890123456789012345678901", "secret key for JWT siging")
	if len(*secretKey) < minSecretKeySize {
		log.Fatalf("SECRET_KEY must be at least %d characters", minSecretKeySize)
	}

	ctx := context.Background()
	dbString := env.GetString("GOOSE_DBSTRING", "user=postgres password=postgres host=localhost port=5432 dbname=ecomm sslmode=disable")

	database, err := db.NewDatabase(ctx, dbString)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	log.Println("successfully connected to database")

	st := storer.NewPostgresStorer(database.GetDB())
	srv := server.NewServer(st)
	hdl := handler.NewHandler(ctx, srv, *secretKey)
	handler.RegisterRoutes(hdl)
	handler.Start(":8080")

}
