package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.smartbet.am/golang/smart-image/internal/service/config"
	"gitlab.smartbet.am/golang/smart-image/internal/service/logger"
	"go.uber.org/fx"
	"strings"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/segmentio/kafka-go"
	"gitlab.smartbet.am/golang/smart-image/ent/image"
	"gitlab.smartbet.am/golang/smart-image/internal/service/db"
	"gitlab.smartbet.am/golang/smart-image/internal/service/fs"
	"gitlab.smartbet.am/golang/smart-image/internal/service/processor"
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

func NewConsumer(config *config.Config, processor *processor.Processor, fs *fs.FS, log *logger.Logger) *Consumer {
	ctx, fn := context.WithCancel(context.Background())

	return &Consumer{
		config:    config,
		processor: processor,
		fs:        fs,
		logger:    log,
		ctx:       ctx,
		cancel:    fn,
	}
}

type Consumer struct {
	config    *config.Config
	reader    *kafka.Reader
	processor *processor.Processor
	fs        *fs.FS
	db        *db.DB
	logger    *logger.Logger
	ctx       context.Context
	cancel    context.CancelFunc
}

func Start(lifecycle fx.Lifecycle, c *Consumer) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go c.start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			c.Close()
			if pool != nil {
				pool.StopWait()
			}
			return nil
		},
	})

}

func (c *Consumer) Close() {
	if reader != nil {
		err := reader.Close()
		if err != nil {
			c.logger.Error("Error closing kafka reader", "error", err)
		}
	}
	c.cancel()
	c.logger.Info("Kafka reader closed")
}

func (c *Consumer) readerInit() (*kafka.Reader, error) {

	_, err := kafka.Dial("tcp", c.config.KafkaBroker)
	if err != nil {
		c.logger.Error("Error connecting to kafka broker", "error", err)
		return nil, err
	}
	c.logger.Info("Connected to kafka broker", " broker", " localhost:9094")

	brokers := strings.Split(c.config.KafkaBroker, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		Topic:          c.config.KafkaTopic,
		Brokers:        brokers,
		GroupID:        fmt.Sprintf("%s-%s", c.config.KafkaTopic, "image"),
		StartOffset:    kafka.LastOffset,
		//StartOffset:    kafka.FirstOffset,
		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			c.logger.Info(fmt.Sprintf(msg, args...))
		}),
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			c.logger.Error(fmt.Sprintf(msg, args...))
		}),
	})
	return r, nil
}

func (c *Consumer) start() (err error) {
	c.logger.Info("Starting consumer")
	pool = workerpool.New(10)

	reader, err = c.readerInit()
	if err != nil {
		c.logger.Error("Error initializing kafka reader", "error", err)
		return err
	}
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.logger.Info("Closing consumer")
				return
			default:
				m, err := reader.ReadMessage(context.Background())
				if err != nil {
					c.logger.Error("Error reading message ", "error ", err)
					break
				}
				c.handleSave(m)
			}

		}
	}()
	return nil
}

func (c *Consumer) handleSave(m kafka.Message) {
	data := m.Value
	var message Message
	err := json.Unmarshal(data, &message)
	c.logger.Info("Message received ", "data ", string(data))

	if err != nil {
		c.logger.Error("Error unmarshalling message", "error", err)
	}

	pool.Submit(func() {
		info, err := c.db.Image.Query().Where(image.UUID(message.UUID)).First(context.Background())
		if err != nil {
			c.logger.Error("Error updating image ", "uuid ", message.UUID, "error", err)
			return
		}
		if info.IsProceed {
			c.logger.Info("Image already processed ", "uuid ", message.UUID)
			return
		}

		imageRaw, err := c.fs.Read(info.TmpURL)
		if err != nil {
			c.logger.Error("Error reading image ", "tmp_url ", info.TmpURL, "error", err)
			return
		}

		var size *processor.Size
		if message.Size != nil {
			size = message.Size
			imageRaw, err = c.processor.Process(
				processor.WithImage(imageRaw),
				processor.WithContentType(info.ContentType),
				processor.WithOperations(processor.ResizeOperation(message.Size.Width, message.Size.Height)),
			)
			if err != nil {
				c.logger.Error("Error processing image", "uuid", message.UUID, "error", err)
				return
			}
		}

		ext, err := processor.ReleaseExtension(info.ContentType)
		if err != nil {
			c.logger.Error("Error getting extension", "content_type", info.ContentType, "error", err)
			return
		}

		url := fmt.Sprintf("/%s/%s/%s.%s", message.Service, message.Type, message.ID, ext)

		err = c.fs.Write(url, imageRaw, info.ContentType)
		if err != nil {
			c.logger.Error("Error writing image", "url", url, "error", err)
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
			c.logger.Error("Error updating image metadata", "uuid", message.UUID, "error", err)
			return
		}
	})

}
