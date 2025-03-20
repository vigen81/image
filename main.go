package main

import (
	"fmt"
	httpServer "github.com/micro/plugins/v5/server/http"
	"gitlab.smartbet.am/golang/smart-image/internal/env_os"
	"gitlab.smartbet.am/golang/smart-image/internal/plugin/ams"
	"gitlab.smartbet.am/golang/smart-image/internal/service/broker"
	"gitlab.smartbet.am/golang/smart-image/internal/service/db"
	"gitlab.smartbet.am/golang/smart-image/internal/web/route"
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
		return
	}

	err = cgf.Load(ams.NewSource(
		ams.WithSecretName(fmt.Sprintf("/%s/imagix", env_os.Env())),
	))

	if err != nil {
		panic(err)
		return
	}

	if err := cgf.Scan(&cnf); err != nil {
		logger.Fatal(err.Error())
	}
	for k, v := range cnf {
		err := os.Setenv(k, fmt.Sprintf("%v", v))
		if err != nil {
			logger.Fatal(err.Error())
		}
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
