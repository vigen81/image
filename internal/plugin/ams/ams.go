package ams

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type Source struct {
	secretName string
}

type Option func(*Source)

func (s *Source) Read() ([]byte, error) {
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
		Name:           aws.String(s.secretName),
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		return nil, err
	}
	raw = []byte(*param.Parameter.Value)

	return raw, nil
}

func NewSource(opts ...Option) *Source {
	s := &Source{}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithSecretName(name string) Option {
	fmt.Printf("secret name: %s \n", name)
	return func(o *Source) {
		o.secretName = name
	}
}
