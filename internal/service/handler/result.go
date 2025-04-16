package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gitlab.smartbet.am/golang/smart-image/internal/service/db"
	"gitlab.smartbet.am/golang/smart-image/internal/service/fs"
	"gitlab.smartbet.am/golang/smart-image/internal/service/processor"
	"net/http"
)

type UploadRequest struct {
	File string `json:"file"`
}
type UploadResponse struct {
	UUID string `json:"uuid"`
	URL  string `json:"url"`
}

type Result struct {
	processor *processor.Processor
	db        *db.DB
	fs        *fs.FS
}

func (r *Result) Upload(c *fiber.Ctx) error {
	var u UploadRequest

	if err := c.BodyParser(&u); err != nil {
		return err
	}

	if u.File == "" {
		return errors.New("file is required")
	}

	if len(u.File) > 6*1024*1024 {
		return errors.New("file size is too large")
	}
	dataUrl := processor.NewDataUrl(u.File)

	if err := dataUrl.Parse(); err != nil {
		return err
	}

	image, err := dataUrl.Decode()

	if err != nil {
		return err
	}
	if processor.ShouldConvertToWebp(dataUrl.ContentType()) {
		image, err = r.processor.Process(
			processor.WithImage(image),
			processor.WithContentType(dataUrl.ContentType()),
			processor.WithOperations(
				processor.ConvertOperation("webp"),
				processor.ResizeOperation(1600),
			))
		if err != nil {
			return err
		}
	}

	id := uuid.New().String()
	tmpUrl := fmt.Sprintf("tmp/%s", id)
	err = r.fs.NewOperation(
		fs.WithFilename(tmpUrl),
		fs.WithFile(image),
		fs.WithContentType(dataUrl.ContentType()),
	).Write()

	if err != nil {
		return err
	}

	err = r.db.Image.Create().
		SetTmpURL(tmpUrl).
		SetUUID(id).
		SetContentType(dataUrl.ContentType()).
		Exec(context.Background())

	if err != nil {
		deleteError := r.fs.NewOperation(fs.WithFilename(tmpUrl)).Delete()
		if deleteError != nil {
			return err
		}
		return err
	}

	return c.Status(http.StatusOK).JSON(UploadResponse{
		UUID: id,
		URL:  tmpUrl,
	})

}

func NewResult(processor *processor.Processor, fs *fs.FS, db *db.DB) *Result {
	return &Result{
		processor: processor,
		db:        db,
		fs:        fs,
	}
}
