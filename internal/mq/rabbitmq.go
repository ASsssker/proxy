package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ASsssker/proxy/internal/config"
	"github.com/ASsssker/proxy/internal/models"
	"github.com/ASsssker/proxy/internal/services"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn      *amqp091.Connection
	ch        *amqp091.Channel
	queueName string
	log       *slog.Logger
}

func NewRabbitMQ(cfg config.Config, log *slog.Logger) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(cfg.RabbitDNS())
	if err != nil {
		return nil, fmt.Errorf("failed to open rabbitMQ connection: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open rabbitMQ channel: %v", err)
	}

	_, err = ch.QueueDeclare(cfg.RabbitTaskQueueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create rabbitMQ queue: %v", err)
	}

	if err := ch.Qos(int(cfg.RequesterWorkersCount), 0, false); err != nil {
		return nil, fmt.Errorf("failed to set rabbitMQ QOS settigns: %v", err)
	}

	return &RabbitMQ{
		conn:      conn,
		ch:        ch,
		queueName: cfg.RabbitTaskQueueName,
		log:       log,
	}, nil
}

func (r *RabbitMQ) SendTask(ctx context.Context, task models.Task) error {
	const op = "rabbitMQ.SendMessage"
	requestID := ctx.Value(services.RequestIDKey).(string)

	log := r.log.With(slog.String("op", op), slog.String(services.RequestIDKey, requestID))
	log.DebugContext(ctx, "start operation")

	msg, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("%s request_id=%s failed to marshal task: %v", op, requestID, err)
	}

	err = r.ch.PublishWithContext(
		ctx,
		"",
		r.queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        msg,
		},
	)

	if err != nil {
		return fmt.Errorf("%s request_id=%s failed to publish task: %v", op, requestID, err)
	}

	log.DebugContext(ctx, "the operation was successfully completed")

	return nil
}

func (r *RabbitMQ) Subscribe(_ context.Context, taskChan chan models.Task) (context.CancelFunc, error) {
	msgChan, err := r.ch.Consume(
		r.queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to cunsume mq: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case msg := <-msgChan:
				r.log.Debug("mq receiver message", slog.String("message_body", string(msg.Body)))

				task := models.Task{}
				if err := json.Unmarshal(msg.Body, &task); err != nil {
					r.log.Error("failed to unmrashelled mq message", slog.String("message_body", string(msg.Body)))
					continue
				}

				select {
				case taskChan <- task:
				case <-ctx.Done():
					close(taskChan)
					return
				}

			case <-ctx.Done():
				close(taskChan)
				return
			}
		}
	}()

	return cancel, nil
}

func (r *RabbitMQ) Close(_ context.Context) error {
	if errChanClose := r.ch.Close(); errChanClose != nil {
		if err := r.conn.Close(); err != nil {
			return fmt.Errorf("failed to close rabbit connection and channel: %v: %v", err, errChanClose)
		}

		return fmt.Errorf("failed to close rabbit channel: %v", errChanClose)
	}

	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("failed to close rabbit connection: %v", err)
	}

	return nil
}
