package services

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ASsssker/proxy/internal/models"
	"github.com/ASsssker/proxy/internal/validation"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAddTask_GoodPath(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		task models.NewTask
	}{
		{
			name: "lower method name",
			ctx:  newContextWithRequestID(),
			task: models.NewTask{
				URL:    "http://example.com",
				Method: "get",
				Headers: map[string]string{
					"custom-header": "custom_value",
				},
				Body: "body"},
		},
		{
			name: "upper method name",
			ctx:  newContextWithRequestID(),
			task: models.NewTask{
				URL:    "http://example.com",
				Method: "GET",
				Headers: map[string]string{
					"custom-header": "custom_value",
				},
				Body: "body"},
		},
	}

	mockSender := new(mockMessageBroker)

	mockProvider := new(mockTaskProvider)

	for _, tt := range tests {
		mockSender.On("Send", tt.ctx.Value(RequestIDKey)).Return(nil)
		mockProvider.On("AddTask", tt.ctx.Value(RequestIDKey)).Return(nil)
	}
	service := newProxyService(mockProvider, mockSender)
	for _, tt := range tests {
		id, err := service.AddTask(tt.ctx, tt.task)
		require.NoError(t, err)
		require.NotPanics(t, func() { uuid.MustParse(id) })
	}
}

func TestAddTask_BadPath(t *testing.T) {
	errUndefTaskProvider := errors.New("undefined task provider error")
	errUndefMsgSender := errors.New("undefined message sender error")

	tests := []struct {
		name            string
		ctx             context.Context
		task            models.NewTask
		errTaskProvider error
		errMsgSender    error
		errExpected     error
	}{
		{
			name: "invalid url",
			ctx:  newContextWithRequestID(),
			task: models.NewTask{
				URL:    "http:dsd//incorrect.",
				Method: "GET",
			},
			errExpected: ErrValidation,
		},
		{
			name: "empty url",
			ctx:  newContextWithRequestID(),
			task: models.NewTask{
				Method: "POST",
			},
			errExpected: ErrValidation,
		},
		{
			name: "not allowed method",
			ctx:  newContextWithRequestID(),
			task: models.NewTask{
				URL:    "http://example.com",
				Method: "PUT",
			},
			errExpected: ErrValidation,
		},
		{
			name: "invalid method",
			ctx:  newContextWithRequestID(),
			task: models.NewTask{
				URL:    "http://example.com",
				Method: "UNDEFINED",
			},
			errExpected: ErrValidation,
		},
		{
			name: "empty method",
			ctx:  newContextWithRequestID(),
			task: models.NewTask{
				Method: "POST",
			},
			errExpected: ErrValidation,
		},
		{
			name: "task provider undefined error",
			ctx:  newContextWithRequestID(),
			task: models.NewTask{
				URL:    "http://example.com",
				Method: "get",
			},
			errTaskProvider: errUndefTaskProvider,
			errExpected:     errUndefTaskProvider,
		},
		{
			name: "message sender undefined error",
			ctx:  newContextWithRequestID(),
			task: models.NewTask{
				URL:    "http://example.com",
				Method: "get",
			},
			errMsgSender: errUndefMsgSender,
			errExpected:  errUndefMsgSender,
		},
	}

	mockSender := new(mockMessageBroker)
	mockProvider := new(mockTaskProvider)
	for _, tt := range tests {
		mockSender.On("Send", tt.ctx.Value(RequestIDKey)).Return(tt.errMsgSender)
		mockProvider.On("AddTask", tt.ctx.Value(RequestIDKey)).Return(tt.errTaskProvider)
	}

	service := newProxyService(mockProvider, mockSender)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := service.AddTask(tt.ctx, tt.task)
			require.Empty(t, id)
			require.ErrorIs(t, err, tt.errExpected)
		})
	}
}

func newProxyService(taskProvider TaskProvider, msgSender MessageSender) ProxyService {
	validator, err := validation.NewValidator()
	if err != nil {
		panic(err)
	}

	return ProxyService{
		log:          slog.New(slog.DiscardHandler),
		taskProvider: taskProvider,
		msgSender:    msgSender,
		validator:    validator,
	}
}

func newContextWithRequestID() context.Context {
	return context.WithValue(context.Background(), RequestIDKey, uuid.NewString())
}
