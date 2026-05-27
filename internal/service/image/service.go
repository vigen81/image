package image

import (
	"context"
	"fmt"

	"gitlab.smartbet.am/golang/smart-image/ent/image"
	"gitlab.smartbet.am/golang/smart-image/internal/service/db"
	"gitlab.smartbet.am/golang/smart-image/internal/service/fs"
	"gitlab.smartbet.am/golang/smart-image/internal/service/logger"
	"gitlab.smartbet.am/golang/smart-image/internal/service/processor"
)

type ProcessingRequest struct {
	UUID    string
	Service string
	Type    string
	ID      string
	Size    *processor.Size
}

type Service struct {
	processor *processor.Processor
	fs        *fs.FS
	db        *db.DB
	logger    *logger.Logger
}

func NewService(db *db.DB, processor *processor.Processor, fs *fs.FS, log *logger.Logger) *Service {
	return &Service{
		processor: processor,
		fs:        fs,
		db:        db,
		logger:    log,
	}
}

func (s *Service) Process(ctx context.Context, req ProcessingRequest) error {
	info, err := s.db.Image.Query().Where(image.UUID(req.UUID)).First(ctx)
	if err != nil {
		s.logger.Error("Error fetching image", "uuid", req.UUID, "error", err)
		return err
	}
	if info.IsProceed {
		s.logger.Info("Image already processed", "uuid", req.UUID)
		return nil
	}

	imageRaw, err := s.fs.Read(info.TmpURL)
	if err != nil {
		s.logger.Error("Error reading image from S3", "tmp_url", info.TmpURL, "error", err)
		return err
	}

	var size *processor.Size
	if req.Size != nil {
		size = req.Size
		imageRaw, err = s.processor.Process(
			processor.WithImage(imageRaw),
			processor.WithContentType(info.ContentType),
			processor.WithOperations(processor.ResizeOperation(req.Size.Width, req.Size.Height)),
		)
		if err != nil {
			s.logger.Error("Error processing image", "uuid", req.UUID, "error", err)
			return err
		}
	} else if processor.ShouldConvertToWebp(info.ContentType) {
		// Auto-convert to webp if no size is specified but it's a convertible image
		imageRaw, err = s.processor.Process(
			processor.WithImage(imageRaw),
			processor.WithContentType(info.ContentType),
			processor.WithOperations(
				processor.ConvertOperation("webp"),
				processor.ResizeOperation(1600), // Default large size
			),
		)
		if err != nil {
			s.logger.Error("Error auto-processing image", "uuid", req.UUID, "error", err)
			return err
		}
	}

	ext, err := processor.ReleaseExtension(info.ContentType)
	if err != nil {
		s.logger.Error("Error getting extension", "content_type", info.ContentType, "error", err)
		return err
	}

	url := fmt.Sprintf("/%s/%s/%s.%s", req.Service, req.Type, req.ID, ext)
	if err := s.fs.Write(url, imageRaw, info.ContentType); err != nil {
		s.logger.Error("Error writing image to S3", "url", url, "error", err)
		return err
	}

	if err := info.Update().
		SetURL(url).
		SetIsProceed(true).
		SetService(req.Service).
		SetType(req.Type).
		SetObjectID(req.ID).
		SetNillableSize(size).
		Exec(ctx); err != nil {
		s.logger.Error("Error updating image metadata", "uuid", req.UUID, "error", err)
		return err
	}

	s.logger.Info("Image processed successfully", "uuid", req.UUID, "url", url)
	return nil
}
