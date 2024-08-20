package ams

import (
	"context"
	"fmt"
	"github.com/Phoenix365-tech/imagix/internal/env_os"
	"os"
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

func WithSecretName(name string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, secretName{}, name)
	}
}
