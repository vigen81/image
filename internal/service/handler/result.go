package handler

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gitlab.smartbet.am/golang/smart-image/ent"
	entimage "gitlab.smartbet.am/golang/smart-image/ent/image"
	"gitlab.smartbet.am/golang/smart-image/internal/service/db"
	"gitlab.smartbet.am/golang/smart-image/internal/service/fs"
	"gitlab.smartbet.am/golang/smart-image/internal/service/image"
	"gitlab.smartbet.am/golang/smart-image/internal/service/logger"
	"gitlab.smartbet.am/golang/smart-image/internal/service/processor"
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
	imgSrv    *image.Service
	logger    *logger.Logger
}

// Upload godoc
// @Summary      Upload an image
// @Description  Accepts a base64 data URL, processes it (webp conversion/resize), stores it, and returns its UUID.
// @Tags         images
// @Accept       json
// @Produce      json
// @Param        request  body      UploadRequest   true  "Upload payload"
// @Success      200      {object}  UploadResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /upload  [post]
func (r *Result) Upload(c *fiber.Ctx) error {
	var req UploadRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if err := validateUploadRequest(req); err != nil {
		return err
	}

	dataURL := processor.NewDataUrl(req.File)
	if err := dataURL.Parse(); err != nil {
		return err
	}

	contentType := dataURL.ContentType()
	id := uuid.New().String()
	tmpPath := fmt.Sprintf("tmp/%s", id)

	fileBytes, err := getFileBytes(r.processor, dataURL, req.File, contentType)
	if err != nil {
		return err
	}

	if err := saveToFS(r.fs, tmpPath, fileBytes, contentType); err != nil {
		return err
	}

	node, err := saveToDB(r.db, id, tmpPath, contentType)
	if err != nil {
		_ = r.fs.NewOperation(fs.WithFilename(tmpPath)).Delete()
		return err
	}

	ctx := context.Background()
	r.logger.Info("inserted", "id", node.ID, "uuid", id, "where", r.db.Probe(ctx))

	// read-back on the same client
	back, rerr := r.db.Image.Query().Where(entimage.UUID(id)).Only(ctx)
	r.logger.Info("readback", "uuid", id, "found", rerr == nil, "err", rerr, "where", r.db.Probe(ctx))
	_ = back

	return c.Status(fiber.StatusOK).JSON(UploadResponse{UUID: id, URL: tmpPath})
}

func validateUploadRequest(req UploadRequest) error {
	if req.File == "" {
		return errors.New("file is required")
	}
	if len(req.File) > 6*1024*1024 {
		return errors.New("file size is too large")
	}
	return nil
}

func getFileBytes(p *processor.Processor, dataURL *processor.DataUrl, rawData, contentType string) ([]byte, error) {
	if processor.IsImage(contentType) {
		imgBytes, err := dataURL.Decode()
		if err != nil {
			return nil, err
		}
		if processor.ShouldConvertToWebp(contentType) {
			imgBytes, err = p.Process(
				processor.WithImage(imgBytes),
				processor.WithContentType(contentType),
				processor.WithOperations(
					processor.ConvertOperation("webp"),
					processor.ResizeOperation(1600),
				),
			)
			if err != nil {
				return nil, err
			}
		}
		return imgBytes, nil
	} else if processor.IsVideo(contentType) {
		parts := strings.SplitN(rawData, ",", 2)
		if len(parts) != 2 {
			return nil, errors.New("invalid data URL")
		}
		return base64.StdEncoding.DecodeString(parts[1])
	}

	return nil, fmt.Errorf("unsupported content type: %s", contentType)
}

func saveToFS(fsClient *fs.FS, path string, data []byte, contentType string) error {
	return fsClient.NewOperation(
		fs.WithFilename(path),
		fs.WithFile(data),
		fs.WithContentType(contentType),
	).Write()
}

func saveToDB(dbClient *db.DB, id, tmpPath, contentType string) (*ent.Image, error) {
	return dbClient.Image.Create().
		SetTmpURL(tmpPath).SetUUID(id).SetContentType(contentType).
		Save(context.Background())
}

func NewResult(processor *processor.Processor, fs *fs.FS, db *db.DB, imgSrv *image.Service, log *logger.Logger) *Result {
	return &Result{
		processor: processor,
		db:        db,
		fs:        fs,
		imgSrv:    imgSrv,
		logger:    log,
	}
}
