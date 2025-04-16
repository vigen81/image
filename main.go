package main

import (
	"gitlab.smartbet.am/golang/smart-image/internal/service/broker"
	"gitlab.smartbet.am/golang/smart-image/internal/service/config"
	"gitlab.smartbet.am/golang/smart-image/internal/service/db"
	"gitlab.smartbet.am/golang/smart-image/internal/service/fs"
	"gitlab.smartbet.am/golang/smart-image/internal/service/handler"
	"gitlab.smartbet.am/golang/smart-image/internal/service/processor"
	"gitlab.smartbet.am/golang/smart-image/internal/web/route"
	"go.uber.org/fx"
)

var (
	serviceName = "smart-image"
	version     = "latest"
)

func main() {

	app := fx.New(
		fx.Supply(serviceName),
		processor.Module,
		fx.Provide(
			config.Provider,
			db.Provider,
			broker.NewConsumer,
			fs.NewFS,
			route.NewApp,
			handler.NewResult,
		),
		fx.Invoke(
			route.StartServer,
			broker.Start,
		),
	)

	app.Run()
}
