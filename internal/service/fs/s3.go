package fs

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jszwec/s3fs"
)

var once sync.Once

type FS struct {
	*s3fs.S3FS
	ctx *s3.S3
}

func (fs *FS) Write(filename string, data []byte, contentType string) error {
	cacheControlHeader := "max-age=600"
	filename = strings.Trim(filename, "/")
	_, err := fs.ctx.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("AWS_BUCKET")),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		Metadata: map[string]*string{
			"Cache-Control": aws.String(cacheControlHeader),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (fs *FS) Delete(filename string) error {
	filename = strings.Trim(filename, "/")
	_, err := fs.ctx.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(os.Getenv("AWS_BUCKET")),
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

var fs *FS

func Fs() *FS {
	once.Do(func() {
		var bucket = os.Getenv("AWS_BUCKET") // "bucket-name
		s, err := session.NewSession(
			&aws.Config{
				Region: aws.String(os.Getenv("AWS_REGION")),
			})
		if err != nil {
			log.Fatal(err)
		}
		ctx := s3.New(s)
		fs = &FS{
			ctx:  ctx,
			S3FS: s3fs.New(ctx, bucket),
		}
	})
	return fs

}
