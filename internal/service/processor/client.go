package processor

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"gitlab.smartbet.am/golang/smart-image/internal/service/config"
	"gitlab.smartbet.am/golang/smart-image/internal/service/logger"
)

const field = "file"

var req *resty.Client

func (c client) create() *resty.Client {
	var baseUrl = fmt.Sprintf("http://%s", c.config.Imaginary)
	logger.Log.Info("Creating new resty client")
	req = resty.New()
	req.SetDebug(true)
	req.SetBaseURL(baseUrl)
	c.req = req
	return c.req
}

type client struct {
	config *config.Config
	req    *resty.Client
}

func NewClient(config *config.Config) *resty.Client {
	c := &client{
		config: config,
	}
	return c.create()
}
