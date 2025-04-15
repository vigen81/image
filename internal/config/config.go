package config

import (
	"encoding/json"
	"fmt"
	"gitlab.smartbet.am/golang/smart-image/internal/plugin/ams"
	"os"
	"sync"
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
	AWS_Bucket  string `json:"aws_bucket"`
	AWS_Region  string `json:"aws_region"`
	AWS_S3Host  string `json:"aws_s3_host"`
}

var config *Config

var once sync.Once

func Run(serviceName string) error {
	var errX error
	once.Do(func() {
		var data []byte

		if os.Getenv("POD_ENV") == "local" {
			data = []byte(mockConfig)
		} else {
			secretName := fmt.Sprintf("/%s/%s", os.Getenv("POD_ENV"), serviceName)
			r := ams.NewSource(ams.WithSecretName(secretName))
			data, errX = r.Read()
			if errX != nil {
				return
			}
		}

		config = &Config{}
		if err := json.Unmarshal(data, config); err != nil {
			errX = err
			return
		}
	})
	fmt.Printf(`Config value %+v`, *config)
	return errX
}

func Get() *Config {
	return config
}
