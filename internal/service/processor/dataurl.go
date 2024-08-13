package processor

import (
	"encoding/base64"
	"errors"
	"strings"
)

type DataUrl struct {
	url         string
	data        string
	contentType string
}

func NewDataUrl(url string) *DataUrl {
	return &DataUrl{
		url: url,
	}
}

func (d *DataUrl) Parse() error {
	if d.url == "" {
		return errors.New("url is required")
	}
	parts := strings.Split(d.url, ",")
	if len(parts) != 2 {
		return errors.New("invalid data url")
	}
	left := parts[0]
	left = strings.ReplaceAll(left, "data:", "")
	left = strings.ReplaceAll(left, ";base64", "")
	d.contentType = left
	d.data = parts[1]
	return nil
}

func (d *DataUrl) ContentType() string {
	return d.contentType
}

func (d *DataUrl) GetData() string {
	return d.data
}

func (d *DataUrl) GetUrl() string {
	return d.url
}

func (d *DataUrl) Decode() (result []byte, err error) {

	result, err = base64.StdEncoding.DecodeString(d.data)
	if err != nil {
		return nil, err
	}
	return result, nil
}
