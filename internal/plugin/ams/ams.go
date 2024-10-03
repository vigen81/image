package ams

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"time"

	"go-micro.dev/v5/config/source"
)

type secretName struct{}

type Source struct {
	options source.Options
}

func (s *Source) Read() (*source.ChangeSet, error) {
	var raw []byte
	var err error
	//if def.IsDev() {
	//raw, err = os.ReadFile(fmt.Sprintf("./config.json"))
	//} else {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1")},
	)
	if err != nil {
		return nil, err
	}
	// Create a new SSM client
	svc := ssm.New(sess)

	// Define the parameter name

	// Get the parameter value
	param, err := svc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(s.options.Context.Value(secretName{}).(string)),
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		return nil, err
	}
	raw = []byte(*param.Parameter.Value)

	cs := &source.ChangeSet{
		Timestamp: time.Now(),
		Format:    s.options.Encoder.String(),
		Source:    s.String(),
		Data:      json.RawMessage(raw),
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

func WithSecretName(name string) source.Option {
	fmt.Printf("secret name: %s \n", name)
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, secretName{}, name)
	}
}
