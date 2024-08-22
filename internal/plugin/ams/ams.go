package ams

import (
	"context"
	"errors"
	"fmt"
	"github.com/Phoenix365-tech/imagix/internal/env_os"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"log"
	"os"
	"time"

	"go-micro.dev/v5/config/source"
)

type LoadType int

const (
	AWS LoadType = iota
	File
)

type secretName struct{}
type loadType struct{}

type Source struct {
	options source.Options
}

func (s *Source) readFromAws() (*source.ChangeSet, error) {

	config, err := config.NewEnvConfig()

	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		return nil, err
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(s.options.Context.Value(secretName{}).(string)),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		// For a list of exceptions thrown, see
		// https://<<{{DocsDomain}}>>/secretsmanager/latest/apireference/API_GetSecretValue.html
		log.Fatal(err.Error())
	}

	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString
}

func (s *Source) Read() (*source.ChangeSet, error) {
	switch s.options.Context.Value(loadType{}).(LoadType) {
	case AWS:
		return readFromAws()
	case File:
		return readFromFile()
	}
	return nil, errors.New("invalid load type")
}
func (s *Source) readFromFile() (*source.ChangeSet, error) {

	var raw []byte
	var err error
	if env_os.IsDev() {
		raw, err = os.ReadFile(fmt.Sprintf("./config.json"))
	} else {
		raw, err = os.ReadFile(fmt.Sprintf("/mnt/secrets/%s", s.options.Context.Value(secretName{})))
	}

	if err != nil {
		return nil, err
	}
	cs := &source.ChangeSet{
		Timestamp: time.Now(),
		Format:    s.options.Encoder.String(),
		Source:    s.String(),
		Data:      raw,
	}
	cs.Checksum = cs.Sum()
	return cs, nil
}

func (s *Source) Write(_ *source.ChangeSet) error {
	return nil
}

func (s *Source) Watch() (source.Watcher, error) {
	return source.NewNoopWatcher()

}

func (s *Source) String() string {
	return "ams"
}

func NewSource(opts ...source.Option) source.Source {
	return &Source{
		options: source.NewOptions(opts...),
	}

}

func WithLoadType(value LoadType) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, loadType{}, value)
	}
}

func WithSecretName(name string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, secretName{}, name)
	}
}
