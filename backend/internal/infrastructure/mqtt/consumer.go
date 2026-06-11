// Package mqtt is the broker-facing adapter: it subscribes to device topics
// and feeds raw messages into the ingest pipeline. Paho handles reconnection;
// OnConnect re-subscribes so a broker restart heals without intervention.
package mqtt

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"github.com/ioss/iot-dashboard/backend/internal/application/ingest"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/config"
)

const (
	topicTelemetry = "devices/+/telemetry"
	topicStatus    = "devices/+/status"
	qosAtLeastOnce = 1
)

// Consumer bridges Mosquitto → ingest.Service.
type Consumer struct {
	client paho.Client
	ingest *ingest.Service
	log    *slog.Logger
}

// NewConsumer configures the client with auto-reconnect + exponential backoff.
func NewConsumer(cfg config.MQTTConfig, svc *ingest.Service, log *slog.Logger) *Consumer {
	c := &Consumer{ingest: svc, log: log}

	opts := paho.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.Host, cfg.Port)).
		SetClientID(cfg.ClientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetCleanSession(false). // broker queues QoS1 messages across restarts
		SetAutoReconnect(true).
		SetMaxReconnectInterval(30 * time.Second).
		SetConnectRetry(true).
		SetConnectRetryInterval(2 * time.Second).
		SetOrderMatters(false). // workers process concurrently anyway
		SetOnConnectHandler(c.onConnect).
		SetConnectionLostHandler(func(_ paho.Client, err error) {
			log.Warn("mqtt connection lost", slog.String("error", err.Error()))
		})

	c.client = paho.NewClient(opts)
	return c
}

// Start connects (retrying until the broker is reachable). Non-blocking after
// the initial connect handshake.
func (c *Consumer) Start() error {
	token := c.client.Connect()
	// With ConnectRetry enabled paho keeps trying; Wait returns on first result.
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqtt connect: %w", token.Error())
	}
	return nil
}

// Stop disconnects cleanly, allowing in-flight handlers 250ms to finish.
func (c *Consumer) Stop() {
	c.client.Disconnect(250)
	c.log.Info("mqtt consumer stopped")
}

// onConnect (re)establishes subscriptions — runs on every (re)connect.
func (c *Consumer) onConnect(client paho.Client) {
	c.log.Info("mqtt connected, subscribing",
		slog.String("telemetry", topicTelemetry),
		slog.String("status", topicStatus),
	)
	if t := client.Subscribe(topicTelemetry, qosAtLeastOnce, c.handle(ingest.KindTelemetry)); t.Wait() && t.Error() != nil {
		c.log.Error("subscribe failed", slog.String("topic", topicTelemetry), slog.Any("error", t.Error()))
	}
	if t := client.Subscribe(topicStatus, qosAtLeastOnce, c.handle(ingest.KindStatus)); t.Wait() && t.Error() != nil {
		c.log.Error("subscribe failed", slog.String("topic", topicStatus), slog.Any("error", t.Error()))
	}
}

// handle adapts a paho callback into an ingest enqueue. The callback does the
// absolute minimum (topic parse + enqueue) — heavy lifting happens in workers.
func (c *Consumer) handle(kind ingest.Kind) paho.MessageHandler {
	return func(_ paho.Client, msg paho.Message) {
		deviceID, ok := deviceIDFromTopic(msg.Topic())
		if !ok {
			return
		}
		c.ingest.Enqueue(ingest.Message{
			DeviceID:   deviceID,
			Kind:       kind,
			Payload:    msg.Payload(),
			ReceivedAt: time.Now().UTC(),
		})
	}
}

// deviceIDFromTopic extracts {id} from devices/{id}/{suffix}.
func deviceIDFromTopic(topic string) (string, bool) {
	parts := strings.Split(topic, "/")
	if len(parts) != 3 || parts[0] != "devices" || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
