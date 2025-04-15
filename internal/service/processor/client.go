package processor

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"gitlab.smartbet.am/golang/smart-image/internal/config"
	"gitlab.smartbet.am/golang/smart-image/internal/logger"
	"sync"
)

const field = "file"

var req *resty.Client

var once sync.Once
var cfg *config.Config

func Req() *resty.Client {
	once.Do(func() {
		cfg = config.Get()
		var baseUrl = fmt.Sprintf("http://%s", cfg.Imaginary)
		logger.Log.Info("Creating new resty client")
		req = resty.New()
		req.SetDebug(true)
		req.SetBaseURL(baseUrl)
	})
	return req
}
