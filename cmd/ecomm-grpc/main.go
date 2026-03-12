package main

import (
	"context"
	"log"
	"net"

	db "github.com/dockyguy56/ecomm/internal/adapters/postgresql/sqlx"
	"github.com/dockyguy56/ecomm/internal/ecomm-grpc/pb"
	"github.com/dockyguy56/ecomm/internal/ecomm-grpc/server"
	"github.com/dockyguy56/ecomm/internal/ecomm-grpc/storer"
	"github.com/dockyguy56/ecomm/internal/env"
	"github.com/ianschenck/envflag"
	"google.golang.org/grpc"
)

func main() {
	var (
		svcAddr = envflag.String("SVC_ADDR", "0.0.0.0:9091", "address where the ecomm-grpc service is listening on")
	)
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

	// register with grpc
	grpcSrv := grpc.NewServer()
	pb.RegisterEcommServer(grpcSrv, srv)

	listerner, err := net.Listen("tcp", *svcAddr)
	if err != nil {
		log.Fatalf("lister failed: %v", err)
	}

	log.Printf("server listening on %s", *svcAddr)
	err = grpcSrv.Serve(listerner)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
