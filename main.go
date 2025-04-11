package main

import (
	"fmt"
	"gitlab.smartbet.am/golang/smart-image/internal/config"
	"gitlab.smartbet.am/golang/smart-image/internal/service/broker"
	"gitlab.smartbet.am/golang/smart-image/internal/service/db"
	"gitlab.smartbet.am/golang/smart-image/internal/web/route"
)

var (
	serviceName = "smart-image"
	version     = "latest"
)

func main() {

	err := config.Run(serviceName)

	hd := route.New()

	_, err = db.Open()
	if err != nil {
		panic(err)
	}
	go func() {
		err = broker.Consume()
		if err != nil {
			fmt.Println(err)
		}
	}()

	err = hd.Listen(":8080")
	if err != nil {
		return
	}
}
