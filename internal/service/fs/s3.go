package fs

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jszwec/s3fs"
	"log"
	"net/http"
	"os"
	"sync"
)

var once sync.Once

type FS struct {
	*s3fs.S3FS
	ctx *s3.S3
}

func (fs *FS) Write(filename string, data []byte) error {

	_, err := fs.ctx.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("AWS_BUCKET")),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(http.DetectContentType(data)),
	})
	if err != nil {
		return err
	}
	return nil
}

func (fs *FS) Delete(filename string) error {
	_, err := fs.ctx.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(os.Getenv("AWS_BUCKET")),
		Key:    aws.String(filename),
	})
	if err != nil {
		return err
	}
	return nil
}

func Fs() *FS {
	var fs *FS
	fs = &FS{}
	once.Do(func() {
		var bucket = os.Getenv("AWS_BUCKET") // "bucket-name
		s, err := session.NewSession(
			&aws.Config{
				Region:      aws.String(os.Getenv("AWS_REGION")),
				Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_KEY"), os.Getenv("AWS_SECRET"), ""),
			})
		if err != nil {
			log.Fatal(err)
		}
		ctx := s3.New(s)
		fs.S3FS = s3fs.New(ctx, bucket)
		fs.ctx = ctx

	})
	return fs

}
