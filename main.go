package main

import (
	"fmt"
	"github.com/Phoenix365-tech/imagix/internal/service/db"
	"github.com/Phoenix365-tech/imagix/internal/web/route"
	"github.com/joho/godotenv"
	httpServer "github.com/micro/plugins/v5/server/http"
	"go-micro.dev/v5"
	"go-micro.dev/v5/logger"
	"go-micro.dev/v5/server"
)

var (
	serviceName = "imagix"
	version     = "latest"
)

func main() {
	// Create service
	err := godotenv.Load()
	if err != nil {
		logger.Fatal(err.Error())
	}
	srv := httpServer.NewServer(
		server.Name(serviceName),
		server.Address(fmt.Sprintf(":%d", 8080)),
	)

	service := micro.NewService(
		micro.Name(serviceName),
		micro.Version(version),
		micro.Server(srv),
		micro.BeforeStart(func() error {
			_, err := db.Open()
			if err != nil {
				return err
			}
			return nil
		}),
	)

	hd := srv.NewHandler(route.New())
	if err := srv.Handle(hd); err != nil {
		logger.Fatal(err.Error())
	}

	// Run service
	if err := service.Run(); err != nil {
		logger.Fatal(err)
	}
}
