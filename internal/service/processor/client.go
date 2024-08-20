package processor

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"go-micro.dev/v5/logger"
	"os"
	"sync"
)

const field = "file"

var req *resty.Client

var once sync.Once

func Req() *resty.Client {
	once.Do(func() {
		var baseUrl = fmt.Sprintf("http://%s", os.Getenv("IMAGINARY"))
		logger.Info("Creating new resty client")
		req = resty.New()
		req.SetDebug(true)
		req.SetBaseURL(baseUrl)
	})
	return req
}
