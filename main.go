package main

import (
	"context"
	"fmt"
	"log"
	"net"

	_ "embed"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	usersvcv1 "github.com/mlukasik-dev/faceit-usersvc/gen/go/faceit/usersvc/v1"
	"github.com/mlukasik-dev/faceit-usersvc/internal/appconfig"
	"github.com/mlukasik-dev/faceit-usersvc/internal/controller"
	"github.com/mlukasik-dev/faceit-usersvc/internal/events"
	"github.com/mlukasik-dev/faceit-usersvc/internal/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//go:embed configs/config.yaml
var configFile []byte

func main() {
	if err := appconfig.Init(configFile); err != nil {
		log.Fatal(err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	client, err := store.Connect(appconfig.AppConfig.Mongodb.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	s := store.New(client)
	if err := s.CreateIndexes(context.Background()); err != nil {
		log.Fatal(err)
	}
	e := events.New()
	ctr := controller.New(s, logger, e)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", appconfig.AppConfig.Port))
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger),
		)),
	)
	usersvcv1.RegisterServiceServer(grpcServer, ctr)
	// setup reflection so evens-cli REPL mode can be used for testing.
	reflection.Register(grpcServer)
	fmt.Printf("Listening at %s\n", lis.Addr().String())
	log.Fatal(grpcServer.Serve(lis))
}
