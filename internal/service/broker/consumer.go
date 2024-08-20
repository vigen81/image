package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Phoenix365-tech/imagix/ent/image"
	"github.com/Phoenix365-tech/imagix/internal/service/db"
	"github.com/Phoenix365-tech/imagix/internal/service/fs"
	"github.com/Phoenix365-tech/imagix/internal/service/processor"
	"github.com/gammazero/workerpool"
	"github.com/segmentio/kafka-go"
	"go-micro.dev/v5/logger"
	"os"
)

var (
	reader *kafka.Reader
	pool   *workerpool.WorkerPool
)

type Message struct {
	UUID    string          `json:"uuid"`
	Size    *processor.Size `json:"size"`
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Service string          `json:"service"`
}

func Consume() (err error) {
	logger.Info("Starting consumer")
	pool = workerpool.New(10)
	_, err = kafka.Dial("tcp", os.Getenv("KAFKA_BROKER"))
	if err != nil {
		return err
	}
	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{os.Getenv("KAFKA_BROKER")},
		Dialer:   kafka.DefaultDialer,
		Topic:    os.Getenv("KAFKA_TOPIC"),
		GroupID:  "main",
		MaxBytes: 10e6, // 10MB
	})

	go func() {
		for {
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				logger.Errorf("Error reading message: %v", err)
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
	logger.Infof("Message: %s", string(data))
	if err != nil {
		logger.Errorf("Error unmarshalling message: %v", err)
	}

	pool.Submit(func() {

		info, err := db.Client().Image.Query().Where(image.UUID(message.UUID)).First(context.Background())
		if err != nil {
			logger.Errorf("Error updating image: %v", err)
			return
		}
		if info.IsProceed {
			logger.Infof("Image already processed: %s", message.UUID)
			err = reader.CommitMessages(context.Background(), m)
			if err != nil {
				logger.Errorf("Error committing message: %v", err)
			}
			return
		}

		imageRaw, err := fs.Fs().Read(info.TmpURL)

		if err != nil {
			logger.Errorf("Error reading image: %v", err)
			return
		}
		var size *processor.Size
		if nil != message.Size {
			size = message.Size
			imageRaw, err = processor.Process(
				processor.WithImage(imageRaw),
				processor.WithContentType(info.ContentType),
				processor.WithOperations(processor.ResizeOperation(message.Size.Width, message.Size.Height)),
			)
			if err != nil {
				logger.Errorf("Error processing image: %v", err)
				return
			}

		}

		ext, err := processor.ReleaseExtension(info.ContentType)

		if err != nil {
			logger.Errorf("Error getting extension: %v", err)
			return
		}

		url := fmt.Sprintf("/%s/%s/%s.%s", message.Service, message.Type, message.ID, ext)

		err = fs.Fs().Write(url, imageRaw, info.ContentType)
		if err != nil {
			logger.Errorf("Error writing image: %v", err)
			return
		}
		err = reader.CommitMessages(context.Background(), m)
		if err != nil {
			logger.Errorf("Error committing message: %v", err)
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
			logger.Errorf("Error updating image: %v", err)
			return
		}
	})

}
