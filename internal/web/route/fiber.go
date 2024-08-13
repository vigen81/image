package route

import (
	"context"
	"errors"
	"fmt"
	"github.com/Phoenix365-tech/imagix/internal/service/db"
	"github.com/Phoenix365-tech/imagix/internal/service/fs"
	"github.com/Phoenix365-tech/imagix/internal/service/processor"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"net/http"
)

type UploadRequest struct {
	File string `json:"file"`
}
type UploadResponse struct {
	UUID string `json:"uuid"`
	URL  string `json:"url"`
}

func routes(app *fiber.App) {

	app.Post("/upload", func(c *fiber.Ctx) error {
		var u UploadRequest

		if err := c.BodyParser(&u); err != nil {
			return err
		}

		if u.File == "" {
			return errors.New("file is required")
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
			image, err = processor.Process(
				processor.WithImage(image),
				processor.WithContentType(dataUrl.ContentType()),
				processor.WithOperations(
					processor.ConvertOperation("webp"),
					processor.ResizeOperation(600),
				))
			if err != nil {
				return err
			}
		}

		id := uuid.New().String()
		tmpUrl := fmt.Sprintf("tmp/%s", id)
		err = fs.NewOperation(
			fs.WithFilename(tmpUrl),
			fs.WithFile(image),
			fs.WithContentType(dataUrl.ContentType()),
		).Write()

		if err != nil {
			return err
		}

		err = db.Client().Image.Create().
			SetTmpURL(tmpUrl).
			SetUUID(id).
			SetContentType(dataUrl.ContentType()).
			Exec(context.Background())

		if err != nil {
			deleteError := fs.NewOperation(fs.WithFilename(tmpUrl)).Delete()
			if deleteError != nil {
				return err
			}
			return err
		}

		return c.Status(http.StatusOK).JSON(UploadResponse{
			UUID: id,
			URL:  tmpUrl,
		})

	})

}

func New() http.HandlerFunc {

	app := fiber.New(fiber.Config{
		Prefork:       false,
		CaseSensitive: false,
		StrictRouting: true,
		ServerHeader:  "Fiber",
		AppName:       "Bat Server 1.0",
	})

	app.Use(cors.New())
	routes(app)

	return adaptor.FiberApp(app)

}
