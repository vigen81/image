package fs

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jszwec/s3fs"
	"gitlab.smartbet.am/golang/smart-image/internal/service/config"
	"io"
	"log"
	"strings"
)

type FS struct {
	*s3fs.S3FS
	ctx    *s3.S3
	config *config.Config
}

func (fs *FS) Write(filename string, data []byte, contentType string) error {
	cacheControlHeader := "max-age=600"
	filename = strings.Trim(filename, "/")
	_, err := fs.ctx.PutObject(&s3.PutObjectInput{
		Bucket:       aws.String(fs.config.AwsBucket),
		Key:          aws.String(filename),
		Body:         bytes.NewReader(data),
		ContentType:  aws.String(contentType),
		CacheControl: aws.String(cacheControlHeader),
	})
	if err != nil {
		return err
	}
	return nil
}

func (fs *FS) Delete(filename string) error {
	filename = strings.Trim(filename, "/")
	_, err := fs.ctx.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(fs.config.AwsBucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return err
	}
	return nil
}

func (fs *FS) Read(filename string) ([]byte, error) {
	filename = strings.Trim(filename, "/")
	f, err := fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	result, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (fs *FS) NewOperation(opts ...Option) Operation {
	op := &operation{
		fs: fs,
	}
	for _, opt := range opts {
		opt(op)
	}
	return op
}

func NewFS(conf *config.Config) *FS {
	var bucket = conf.AwsBucket
	s, err := session.NewSession(
		&aws.Config{
			Endpoint:         aws.String(conf.AwsS3host),
			Region:           aws.String(conf.AwsRegion),
			S3ForcePathStyle: aws.Bool(true),
			DisableSSL:       aws.Bool(true), //delete for prod
		})
	if err != nil {
		log.Fatal(err)
	}
	ctx := s3.New(s)
	fs := &FS{
		ctx:    ctx,
		config: conf,
		S3FS:   s3fs.New(ctx, bucket),
	}
	return fs
}
