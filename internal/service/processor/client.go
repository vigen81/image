package processor

import (
	"github.com/go-resty/resty/v2"
	"go-micro.dev/v5/logger"
	"sync"
)

const field = "file"
const baseUrl = "http://localhost:9000"

var req *resty.Client

var once sync.Once

func Req() *resty.Client {
	once.Do(func() {
		logger.Info("Creating new resty client")
		req = resty.New()
		req.SetDebug(true)
		req.SetBaseURL(baseUrl)
	})
	return req
}
