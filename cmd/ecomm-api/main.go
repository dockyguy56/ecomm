package main

import (
	"context"
	"log"

	"github.com/dockyguy56/ecomm/internal/ecomm-api/handler"
	"github.com/dockyguy56/ecomm/internal/ecomm-grpc/pb"
	"github.com/ianschenck/envflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const minSecretKeySize = 32

func main() {
	// Load environment variables
	var (
		secretKey = envflag.String("SECRET_KEY", "01234567890123456789012345678901", "secret key for JWT siging")
		svcAddr   = envflag.String("GRPC_SVC_ADDR", "0.0.0.0:9091", "address where the ecomm-grpc service is listening on")
	)
	if len(*secretKey) < minSecretKeySize {
		log.Fatalf("SECRET_KEY must be at least %d characters", minSecretKeySize)
	}

	ctx := context.Background()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(*svcAddr, opts...)
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	log.Printf("Connected to client on %v", *svcAddr)

	client := pb.NewEcommClient(conn)
	hdl := handler.NewHandler(ctx, client, *secretKey)
	handler.RegisterRoutes(hdl)
	handler.Start(":8080")

}
