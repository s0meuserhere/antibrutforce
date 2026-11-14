package grpcserver

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	antibfv1 "github.com/s0meuserhere/antibrutforce/gen/antibf/v1"
	"github.com/s0meuserhere/antibrutforce/internal/service/guard"
)

// New возвращает сервер gRPC с ручками анти-брутфорса.
func New(service *guard.Service, logger *slog.Logger, requestTimeout time.Duration) *grpc.Server {
	server := grpc.NewServer(grpc.UnaryInterceptor(func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		next grpc.UnaryHandler,
	) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		return next(ctx, req)
	}))

	antibfv1.RegisterAntiBruteforceServiceServer(server, &handler{service: service, logger: logger})

	reflection.Register(server)

	return server
}
