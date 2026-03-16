package broker

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/aws/aws-msk-iam-sasl-signer-go/signer"
	"gitlab.smartbet.am/golang/smart-image/ent/image"
	"gitlab.smartbet.am/golang/smart-image/internal/service/config"
	"gitlab.smartbet.am/golang/smart-image/internal/service/db"
	"gitlab.smartbet.am/golang/smart-image/internal/service/fs"
	"gitlab.smartbet.am/golang/smart-image/internal/service/logger"
	"gitlab.smartbet.am/golang/smart-image/internal/service/processor"
	"go.uber.org/fx"
)

type MSKAccessTokenProvider struct {
	Region string
}

func (m *MSKAccessTokenProvider) Token() (*sarama.AccessToken, error) {
	token, _, err := signer.GenerateAuthToken(context.TODO(), m.Region)
	return &sarama.AccessToken{Token: token}, err
}

type Message struct {
	UUID    string          `json:"uuid"`
	Size    *processor.Size `json:"size"`
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Service string          `json:"service"`
}

type Consumer struct {
	config     *config.Config
	subscriber message.Subscriber
	processor  *processor.Processor
	fs         *fs.FS
	db         *db.DB
	logger     *logger.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewConsumer(cfg *config.Config, db *db.DB, processor *processor.Processor, fs *fs.FS, log *logger.Logger) (*Consumer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Verify broker connectivity first
	for _, broker := range cfg.KafkaBrokers() {
		conn, err := net.DialTimeout("tcp", broker, 5*time.Second)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to connect to kafka broker %s: %w", broker, err)
		}
		conn.Close()
		log.Info("Connected to broker", "broker", broker)
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_0_0
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	if os.Getenv("POD_ENV") != "local" {
		saramaConfig.Net.TLS.Enable = true
		saramaConfig.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: true,
		}
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.Mechanism = sarama.SASLTypeOAuth
		saramaConfig.Net.SASL.TokenProvider = &MSKAccessTokenProvider{Region: "eu-central-1"}
	}

	subscriberConfig := kafka.SubscriberConfig{
		Brokers:               cfg.KafkaBrokers(),
		Unmarshaler:           kafka.DefaultMarshaler{},
		ConsumerGroup:         fmt.Sprintf("%s-image", cfg.KafkaTopic),
		OverwriteSaramaConfig: saramaConfig,
	}

	wmLogger := watermill.NewStdLogger(false, false)
	subscriber, err := kafka.NewSubscriber(subscriberConfig, wmLogger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create kafka subscriber: %w", err)
	}

	return &Consumer{
		config:     cfg,
		subscriber: subscriber,
		processor:  processor,
		fs:         fs,
		db:         db,
		logger:     log,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

func Start(lifecycle fx.Lifecycle, c *Consumer) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go c.start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			c.Close()
			return nil
		},
	})
}

func (c *Consumer) Close() {
	c.cancel()
	if err := c.subscriber.Close(); err != nil {
		c.logger.Error("Error closing kafka subscriber", "error", err)
	}
	c.logger.Info("Kafka subscriber closed")
}

func (c *Consumer) start() {
	c.logger.Info("Starting consumer")

	messages, err := c.subscriber.Subscribe(c.ctx, c.config.KafkaTopic)
	if err != nil {
		c.logger.Error("Failed to subscribe to topic", "topic", c.config.KafkaTopic, "error", err)
		return
	}

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("Closing consumer")
			return
		case msg, ok := <-messages:
			if !ok {
				c.logger.Info("Messages channel closed")
				return
			}
			c.handleSave(msg)
		}
	}
}

func (c *Consumer) handleSave(msg *message.Message) {
	defer msg.Ack() // Ack even on error to avoid infinite redelivery; use Nack() if you want retry

	var m Message
	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		c.logger.Error("Error unmarshalling message", "error", err)
		return
	}
	c.logger.Info("Message received", "uuid", m.UUID)

	info, err := c.db.Image.Query().Where(image.UUID(m.UUID)).First(c.ctx)
	if err != nil {
		c.logger.Error("Error fetching image", "uuid", m.UUID, "error", err)
		return
	}
	if info.IsProceed {
		c.logger.Info("Image already processed", "uuid", m.UUID)
		return
	}

	imageRaw, err := c.fs.Read(info.TmpURL)
	if err != nil {
		c.logger.Error("Error reading image from S3", "tmp_url", info.TmpURL, "error", err)
		return
	}

	var size *processor.Size
	if m.Size != nil {
		size = m.Size
		imageRaw, err = c.processor.Process(
			processor.WithImage(imageRaw),
			processor.WithContentType(info.ContentType),
			processor.WithOperations(processor.ResizeOperation(m.Size.Width, m.Size.Height)),
		)
		if err != nil {
			c.logger.Error("Error processing image", "uuid", m.UUID, "error", err)
			return
		}
	}

	ext, err := processor.ReleaseExtension(info.ContentType)
	if err != nil {
		c.logger.Error("Error getting extension", "content_type", info.ContentType, "error", err)
		return
	}

	url := fmt.Sprintf("/%s/%s/%s.%s", m.Service, m.Type, m.ID, ext)
	if err := c.fs.Write(url, imageRaw, info.ContentType); err != nil {
		c.logger.Error("Error writing image to S3", "url", url, "error", err)
		return
	}

	if err := info.Update().
		SetURL(url).
		SetIsProceed(true).
		SetService(m.Service).
		SetType(m.Type).
		SetObjectID(m.ID).
		SetNillableSize(size).
		Exec(c.ctx); err != nil {
		c.logger.Error("Error updating image metadata", "uuid", m.UUID, "error", err)
		return
	}

	c.logger.Info("Image processed successfully", "uuid", m.UUID, "url", url)
}
