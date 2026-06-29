package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gitlab.smartbet.am/golang/smart-image/internal/plugin/ams"
	"gitlab.smartbet.am/golang/smart-image/internal/service/logger"
	"go.uber.org/fx"
)

const mockConfig = `{
"db_port": "3306",
	"db_host": "localhost",
	"db_user": "root",
	"db_password": "123456!",
	"db_name": "smart_image",
	"api_key": "mock-secret-key",
	"kafka_topic": "smart_image",
	"kafka_broker": "localhost:9094",
	"imaginary": "localhost:9000",
	"aws_bucket": "smart-image",
	"aws_region": "us-east-1",
	"aws_s3_host": "127.0.0.1:4566"
}`

type Config struct {
	DBPort      string `json:"db_port"`
	DBHost      string `json:"db_host"`
	DBUser      string `json:"db_user"`
	DBPassword  string `json:"db_password"`
	DBName      string `json:"db_name"`
	APIKey      string `json:"api_key"`
	KafkaTopic  string `json:"kafka_topic"`
	KafkaBroker string `json:"kafka_broker"`
	Imaginary   string `json:"imaginary"`
	AwsBucket   string `json:"aws_bucket"`
	AwsRegion   string `json:"aws_region"`
	AwsS3host   string `json:"aws_s3_host"`
}

func (cnf *Config) run(serviceName string) error {
	var data []byte
	var err error
	if os.Getenv("POD_ENV") == "local" {
		data = []byte(mockConfig)
	} else {
		secretName := fmt.Sprintf("/%s/%s", os.Getenv("POD_ENV"), serviceName)
		r := ams.NewSource(ams.WithSecretName(secretName))
		data, err = r.Read()
		if err != nil {
			return err
		}
	}

	if err := json.Unmarshal(data, cnf); err != nil {
		return err
	}
	return err
}

func Provider(lifecycle fx.Lifecycle, serviceName string, log *logger.Logger) *Config {
	c := &Config{}
	//serviceName := "smart-image"
	err := c.run(serviceName)
	log.Warn("config11 ", c)
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return err
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
	return c

}

func (c *Config) KafkaBrokers() []string {
	brokers := strings.Split(c.KafkaBroker, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}
	return brokers
}
