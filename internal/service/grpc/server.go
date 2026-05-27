package grpc

import (
	"context"
	"fmt"

	"gitlab.smartbet.am/golang/sdk/connector/config"
	"gitlab.smartbet.am/golang/sdk/connector/server"
	smartimagev1 "gitlab.smartbet.am/golang/smart-image/pb/v1/smart-image"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type Server struct {
	smartimagev1.UnimplementedSmartImageServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) UploadImage(ctx context.Context, req *smartimagev1.UploadRequest) (*smartimagev1.Empty, error) {
	fmt.Printf("Received UploadImage request for UUID: %s\n", req.GetUuid())
	// Implementation will go here later if needed
	return &smartimagev1.Empty{}, nil
}

func StartGRPCServer(lc fx.Lifecycle, srv *Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			cfg := &config.ServerConfig{
				Address: ":50050",
			}
			s := server.NewServer(cfg)
			s.RegisterService(func(grpcSrv *grpc.Server) {
				smartimagev1.RegisterSmartImageServiceServer(grpcSrv, srv)
			})

			go func() {
				if err := s.Start(); err != nil {
					fmt.Printf("gRPC server error: %v\n", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
