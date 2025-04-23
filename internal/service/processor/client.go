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
	c.logger.Info("Creating new resty client")
	req = resty.New()
	req.SetDebug(true)
	req.SetBaseURL(baseUrl)
	c.req = req
	return c.req
}

type client struct {
	config *config.Config
	req    *resty.Client
	logger *logger.Logger
}

func NewClient(config *config.Config, logger *logger.Logger) *resty.Client {
	c := &client{
		config: config,
		logger: logger,
	}
	return c.create()
}
