package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ASsssker/proxy/internal/config"
	"github.com/ASsssker/proxy/internal/models"
	"github.com/ASsssker/proxy/internal/services"
	"github.com/nats-io/nats.go"
)

type NatsMQ struct {
	conn      *nats.Conn
	queueName string
	log       *slog.Logger
}

func NewNatsMQ(cfg config.Config, log *slog.Logger) (*NatsMQ, error) {
	nc, err := nats.Connect(cfg.NatsDNS())
	if err != nil {
		return nil, fmt.Errorf("failed to open nats connection: %v", err)
	}

	return &NatsMQ{
			log:       log,
			conn:      nc,
			queueName: cfg.NatsTaskQueueName},
		nil
}

func (n *NatsMQ) SendTask(ctx context.Context, task models.Task) error {
	const op = "nats.SendMessage"
	requestID := ctx.Value(services.RequestIDKey).(string)

	log := n.log.With(slog.String("op", op), slog.String(services.RequestIDKey, requestID))
	log.DebugContext(ctx, "start operation")

	msg, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("%s request_id=%s failed to marshal task: %v", op, requestID, err)
	}

	if err := n.conn.Publish(n.queueName, msg); err != nil {
		return fmt.Errorf("%s request_id=%s failed to publish task: %v", op, requestID, err)
	}

	log.DebugContext(ctx, "the operation was successfully completed")

	return nil

}

func (n *NatsMQ) Subscribe(_ context.Context, taskChan chan models.Task) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())

	sub, err := n.conn.Subscribe(n.queueName, func(msg *nats.Msg) {
		n.log.Debug("mq receiver message", slog.String("message_body", string(msg.Data)))

		task := models.Task{}
		if err := json.Unmarshal(msg.Data, &task); err != nil {
			n.log.Error("failed to unmrashelled mq message", slog.String("message_body", string(msg.Data)))
			return
		}

		select {
		case taskChan <- task:
		case <-ctx.Done():
			close(taskChan)
		}
	})

	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to subcribe queue: %v", err)
	}

	go func() {
		<-ctx.Done()
		if err := sub.Unsubscribe(); err != nil {
			n.log.Error("failed to unsubscribe queue")
		}
	}()

	return cancel, nil
}

func (n *NatsMQ) Close(ctx context.Context) error {
	n.conn.Close()
	return nil
}
