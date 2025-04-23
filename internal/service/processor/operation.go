package processor

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
)

type (
	Size struct {
		Width  int `json:"width,omitempty"`
		Height int `json:"height,omitempty"`
	}
	OperationType string

	OperationData struct {
		Operation string      `json:"operation"`
		Params    interface{} `json:"params"`
	}

	Operation interface {
		GetData() OperationData
		Operation() OperationType
	}

	resizeOperation struct {
		Size
	}
	convertOperation struct {
		Type string `json:"type"`
	}
	Option func(data *pipeline)

	pipeline struct {
		operations  []Operation
		image       []byte
		contentType string
		client      *resty.Client
	}
)

const (
	TypeResize  OperationType = "resize"
	TypeConvert OperationType = "convert"
)

func (r *resizeOperation) GetData() OperationData {
	return OperationData{
		Operation: string(TypeResize),
		Params:    r.Size,
	}
}

func (r *resizeOperation) Operation() OperationType {
	return TypeResize
}

func ResizeOperation(params ...int) Operation {
	var s = Size{}
	if len(params) > 0 {
		s.Width = params[0]
	}
	if len(params) > 1 {
		s.Height = params[1]
	}
	return &resizeOperation{
		Size: s,
	}
}

func (c *convertOperation) Operation() OperationType {
	return TypeConvert
}

func (c *convertOperation) GetData() OperationData {
	return OperationData{
		Operation: string(TypeConvert),
		Params:    c,
	}
}

func ConvertOperation(t string) Operation {
	return &convertOperation{
		Type: t,
	}
}

func WithOperations(operations ...Operation) Option {
	return func(data *pipeline) {
		data.operations = append(data.operations, operations...)
	}
}
func WithImage(image []byte) Option {
	return func(data *pipeline) {
		data.image = image
	}
}

func WithContentType(contentType string) Option {
	return func(data *pipeline) {
		data.contentType = contentType
	}
}

func (p *pipeline) execute() ([]byte, error) {

	if false == ShouldConvertToWebp(p.contentType) {
		return p.image, nil
	}
	r := p.client.R()
	r.SetMultipartField(field, "file", p.contentType, bytes.NewReader(p.image))
	var data []OperationData
	for _, op := range p.operations {
		data = append(data, op.GetData())
	}
	operationsRaw, err := json.Marshal(data)
	r.SetQueryParam("operations", string(operationsRaw))

	if err != nil {
		return nil, err
	}

	resp, err := r.Post("/pipeline")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.String())

	}
	return resp.Body(), nil
}

type Processor struct {
	client *resty.Client
}

func NewProcessor(client *resty.Client) *Processor {
	return &Processor{
		client: client,
	}
}

func (processor *Processor) Process(opts ...Option) ([]byte, error) {
	p := &pipeline{
		client: processor.client,
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.image == nil {
		return nil, errors.New("image is required")
	}
	if p.contentType == "" {
		return nil, errors.New("content type is required")
	}
	return p.execute()
}
