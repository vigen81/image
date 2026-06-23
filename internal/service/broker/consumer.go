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
	"gitlab.smartbet.am/golang/smart-image/ent"
	"gitlab.smartbet.am/golang/smart-image/internal/service/config"
	srv "gitlab.smartbet.am/golang/smart-image/internal/service/image"
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
	imgSrv     *srv.Service
	logger     *logger.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewConsumer(cfg *config.Config, imgSrv *srv.Service, log *logger.Logger) (*Consumer, error) {
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
		imgSrv:     imgSrv,
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
	var m Message
	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		c.logger.Error("Error unmarshalling message", "error", err)
		msg.Ack() // malformed, never retryable
		return
	}

	c.logger.Info("Message received", "uuid", m.UUID)

	err := c.imgSrv.Process(c.ctx, srv.ProcessingRequest{
		UUID: m.UUID, Service: m.Service, Type: m.Type, ID: m.ID, Size: m.Size,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			// Row isn't there and won't appear by retrying — drop, don't loop.
			c.logger.Error("image row missing, dropping message", "uuid", m.UUID)
			msg.Ack()
			return
		}
		// Genuine transient (connection, imaginary, S3) — redeliver.
		c.logger.Error("Error processing image via service", "uuid", m.UUID, "error", err)
		msg.Nack()
		return
	}
	msg.Ack()
}
