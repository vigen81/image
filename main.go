package main

import (
	"fmt"
	"github.com/Phoenix365-tech/imagix/internal/env_os"
	"github.com/Phoenix365-tech/imagix/internal/plugin/ams"
	"github.com/Phoenix365-tech/imagix/internal/service/broker"
	"github.com/Phoenix365-tech/imagix/internal/service/db"
	"github.com/Phoenix365-tech/imagix/internal/web/route"
	httpServer "github.com/micro/plugins/v5/server/http"
	"go-micro.dev/v5"
	"go-micro.dev/v5/config"
	"go-micro.dev/v5/logger"
	"go-micro.dev/v5/server"
	"os"
)

var (
	serviceName = "imagix"
	version     = "latest"
)

func main() {
	// Create service

	var cnf map[string]interface{}
	env_os.SetEnv(os.Getenv("PHOENIX365_ENVIRONMENT"))
	cgf, err := config.NewConfig()

	if err != nil {
		logger.Fatal(err.Error())
	}

	cgf.Load(ams.NewSource(
		ams.WithSecretName(fmt.Sprintf("/%s/imagix", env_os.Env())),
	))

	if err := cgf.Scan(&cnf); err != nil {

		logger.Fatal(err.Error())
	}
	for k, v := range cnf {
		os.Setenv(k, fmt.Sprintf("%v", v))
	}

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
			err = broker.Consume()
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
