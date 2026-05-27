package grpc

import (
	"context"
	"fmt"
	"net"

	smartimagev1 "gitlab.smartbet.am/golang/smart-image/pb/v1/smart-image"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
			lis, err := net.Listen("tcp", ":50050")
			if err != nil {
				return err
			}
			s := grpc.NewServer()
			smartimagev1.RegisterSmartImageServiceServer(s, srv)
			reflection.Register(s)

			go func() {
				if err := s.Serve(lis); err != nil {
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
