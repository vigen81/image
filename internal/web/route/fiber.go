package route

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/Phoenix365-tech/imagix/internal/service/db"
	"github.com/Phoenix365-tech/imagix/internal/service/fs"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/google/uuid"
	"go-micro.dev/v5/errors"
	"net/http"
	"strings"
)

type DataUrl struct {
	url      string
	data     string
	mimeType string
}

func NewDataUrl(url string) *DataUrl {
	return &DataUrl{
		url: url,
	}
}

func (d *DataUrl) Parse() error {
	if d.url == "" {
		return errors.New("400", "url is required", 400)
	}
	parts := strings.Split(d.url, ",")
	if len(parts) != 2 {
		return errors.New("400", "invalid data url", 400)
	}
	left := parts[0]
	left = strings.ReplaceAll(left, "data:", "")
	left = strings.ReplaceAll(left, ";base64", "")
	d.mimeType = left
	d.data = parts[1]
	return nil
}

func (d *DataUrl) GetMimeType() string {
	return d.mimeType
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

type UploadRequest struct {
	File string `json:"file"`
}

func routes(app *fiber.App) {

	app.Post("/upload", func(c *fiber.Ctx) error {
		var u UploadRequest
		if err := c.BodyParser(&u); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if u.File == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "file is required",
			})
		}

		dataUrl := NewDataUrl(u.File)
		if err := dataUrl.Parse(); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		image, err := dataUrl.Decode()

		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		tmpName := uuid.New().String()
		tmpUrl := fmt.Sprintf("/tmp/%s", tmpName)

		r := resty.New()
		req := r.NewRequest()
		req.SetMultipartField("file", tmpUrl, "image/jpeg", bytes.NewReader(image))

		resp, err := req.Post("http://localhost:9000/resize?width=500&height=400&type=jpeg")
		if err != nil {
			return err
		}

		content := resp.Body()
		//content, err := io.ReadAll(rb)
		if err != nil {
			return err
		}
		err = fs.Fs().Write(tmpUrl, content)

		if err != nil {
			return err
		}

		if resp.StatusCode() != 200 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "failed to upload",
			})
		}

		err = db.Client().Image.Create().
			SetTmpURL(tmpUrl).
			SetUUID(tmpName).
			SetContentType(dataUrl.GetMimeType()).
			Exec(context.Background())

		if err != nil {
			deleteError := fs.Fs().Delete(tmpUrl)
			if deleteError != nil {
				return err
			}
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"message": "file uploaded",
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

	routes(app)

	return adaptor.FiberApp(app)

}
