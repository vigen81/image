package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.smartbet.am/golang/smart-image/internal/config"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/segmentio/kafka-go"
	"gitlab.smartbet.am/golang/smart-image/ent/image"
	"gitlab.smartbet.am/golang/smart-image/internal/logger"
	"gitlab.smartbet.am/golang/smart-image/internal/service/db"
	"gitlab.smartbet.am/golang/smart-image/internal/service/fs"
	"gitlab.smartbet.am/golang/smart-image/internal/service/processor"
)

var (
	reader *kafka.Reader
	pool   *workerpool.WorkerPool
	lcfg   *config.Config
)

type Message struct {
	UUID    string          `json:"uuid"`
	Size    *processor.Size `json:"size"`
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Service string          `json:"service"`
}

func Reader() *kafka.Reader {
	return readerProd()

}

func readerProd() *kafka.Reader {
	lcfg = config.Get()
	//_, err := awscfg.LoadDefaultConfig(context.TODO())

	//addrs := strings.Split(lcfg.KafkaBroker, ",")

	return kafka.NewReader(kafka.ReaderConfig{
		Topic:          lcfg.KafkaTopic,
		Brokers:        []string{"localhost:9092"},
		CommitInterval: 1 * time.Second,
		MaxBytes:       10e6,
		GroupID:        "smart_image_group_main",
		StartOffset:    kafka.FirstOffset,
	})
}

func Consume() (err error) {
	logger.Log.Info("Starting consumer")
	pool = workerpool.New(10)

	reader = Reader()

	go func() {
		for {
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				logger.Log.Error("Error reading message ", "error ", err)
				break
			}
			handleSave(m)
		}
	}()
	return nil
}

func handleSave(m kafka.Message) {
	data := m.Value
	var message Message
	err := json.Unmarshal(data, &message)
	logger.Log.Info("Message received", "data", string(data))

	if err != nil {
		logger.Log.Error("Error unmarshalling message", "error", err)
	}

	pool.Submit(func() {
		info, err := db.Client().Image.Query().Where(image.UUID(message.UUID)).First(context.Background())
		if err != nil {
			logger.Log.Error("Error updating image", "uuid", message.UUID, "error", err)
			return
		}
		if info.IsProceed {
			logger.Log.Info("Image already processed", "uuid", message.UUID)
			return
		}

		imageRaw, err := fs.Fs().Read(info.TmpURL)
		if err != nil {
			logger.Log.Error("Error reading image", "tmp_url", info.TmpURL, "error", err)
			return
		}

		var size *processor.Size
		if message.Size != nil {
			size = message.Size
			imageRaw, err = processor.Process(
				processor.WithImage(imageRaw),
				processor.WithContentType(info.ContentType),
				processor.WithOperations(processor.ResizeOperation(message.Size.Width, message.Size.Height)),
			)
			if err != nil {
				logger.Log.Error("Error processing image", "uuid", message.UUID, "error", err)
				return
			}
		}

		ext, err := processor.ReleaseExtension(info.ContentType)
		if err != nil {
			logger.Log.Error("Error getting extension", "content_type", info.ContentType, "error", err)
			return
		}

		url := fmt.Sprintf("/%s/%s/%s.%s", message.Service, message.Type, message.ID, ext)

		err = fs.Fs().Write(url, imageRaw, info.ContentType)
		if err != nil {
			logger.Log.Error("Error writing image", "url", url, "error", err)
			return
		}

		err = info.Update().SetURL(url).
			SetIsProceed(true).
			SetService(message.Service).
			SetType(message.Type).
			SetObjectID(message.ID).
			SetNillableSize(size).
			Exec(context.Background())
		if err != nil {
			logger.Log.Error("Error updating image metadata", "uuid", message.UUID, "error", err)
			return
		}
	})

}
