package grpc

import (
	"context"
	"fmt"

	"gitlab.smartbet.am/golang/sdk/connector/config"
	"gitlab.smartbet.am/golang/sdk/connector/server"
	srv "gitlab.smartbet.am/golang/smart-image/internal/service/image"
	"gitlab.smartbet.am/golang/smart-image/internal/service/logger"
	"gitlab.smartbet.am/golang/smart-image/internal/service/processor"
	smartimagev1 "gitlab.smartbet.am/golang/smart-image/pb/v1/smart-image"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type Server struct {
	smartimagev1.UnimplementedSmartImageServiceServer
	imgSrv *srv.Service
	logger *logger.Logger
}

func NewServer(imgSrv *srv.Service, log *logger.Logger) *Server {
	return &Server{
		imgSrv: imgSrv,
		logger: log,
	}
}

func (s *Server) UploadImage(ctx context.Context, req *smartimagev1.UploadRequest) (*smartimagev1.Empty, error) {
	fmt.Printf("Received UploadImage request for UUID: %s\n", req.GetUuid())

	var size *processor.Size
	if req.GetSize() != nil {
		size = &processor.Size{
			Width:  int(req.GetSize().GetWidth()),
			Height: int(req.GetSize().GetHeight()),
		}
	}

	err := s.imgSrv.Process(ctx, srv.ProcessingRequest{
		UUID:    req.GetUuid(),
		Service: req.GetService(),
		Type:    req.GetType(),
		ID:      req.GetId(),
		Size:    size,
	})
	if err != nil {
		s.logger.Error("Error processing image via service", "uuid", req.GetUuid(), "error", err)
		return nil, err
	}

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
