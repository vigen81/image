package fs

import "io"

type (
	Operation interface {
		Write() error
		Delete() error
		Read() ([]byte, error)
	}
	operation struct {
		filename    string
		file        []byte
		contentType string
	}
	Option func(data *operation)
)

func NewOperation(opts ...Option) Operation {
	op := &operation{}
	for _, opt := range opts {
		opt(op)
	}
	return op
}

func WithFilename(filename string) Option {
	return func(data *operation) {
		data.filename = filename
	}
}

func WithFile(file []byte) Option {
	return func(data *operation) {
		data.file = file
	}
}

func WithContentType(contentType string) Option {
	return func(data *operation) {
		data.contentType = contentType
	}
}

func (op *operation) Read() ([]byte, error) {
	f, err := Fs().Open(op.filename)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	result, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (op *operation) Write() error {
	return Fs().Write(op.filename, op.file, op.contentType)
}

func (op *operation) Delete() error {
	return Fs().Delete(op.filename)
}
