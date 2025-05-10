package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ASsssker/proxy/internal/models"
	"github.com/ASsssker/proxy/internal/services"
	"github.com/gin-gonic/gin"
)

type ProxyService interface {
	AddTask(ctx context.Context, newTask models.NewTask) (string, error)
	GetTaskInfo(ctx context.Context, taskID string) (models.TaskInfo, error)
}

type Handler struct {
	log          *slog.Logger
	proxyService ProxyService
}

func Register(server gin.IRouter, log *slog.Logger, proxyService ProxyService) {
	RegisterHandlers(server, Handler{log: log, proxyService: proxyService})
}

func (h Handler) AddTask(ctx *gin.Context) {
	var newTask models.NewTask
	if err := ctx.ShouldBindBodyWithJSON(&newTask); err != nil {
		h.handlingError(ctx, err)
		return
	}

	taskID, err := h.proxyService.AddTask(ctx, newTask)
	if err != nil {
		h.handlingError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"id": taskID})
}

func (h Handler) GetTaskResult(ctx *gin.Context, id string) {
	taskInfo, err := h.proxyService.GetTaskInfo(ctx, id)
	if err != nil {
		h.handlingError(ctx, err)
	}

	ctx.JSON(http.StatusOK, taskInfo)
}

type errorResponse struct {
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

func newErrorResponse(errorCode int, description string) errorResponse {
	return errorResponse{ErrorCode: errorCode, Description: description}
}

func (h Handler) handlingError(ctx *gin.Context, err error) {
	requestID, ok := ctx.Get(services.RequestIDKey)
	if !ok {
		h.log.Error("context not contains " + services.RequestIDKey)
		requestID = "not found"
	}

	var requestIDWithStr string
	requestIDWithStr, _ = requestID.(string)

	switch {
	case errors.Is(err, services.ErrTaskNotFound):
		h.log.Debug("task not found", slog.String(services.RequestIDKey, requestIDWithStr))
		ctx.JSON(http.StatusNotFound, newErrorResponse(http.StatusNotFound, "task not found"))

	case errors.Is(err, services.ErrValidation):
		h.log.Debug("task not found", slog.String(services.RequestIDKey, requestIDWithStr))
		// TODO: добавлять подробную информациб о причине ошибки
		ctx.JSON(http.StatusBadRequest, newErrorResponse(http.StatusBadRequest, "validation error"))

	default:
		h.log.Error("undefined error", slog.String(services.RequestIDKey, requestIDWithStr),
			slog.String("error", err.Error()))

		ctx.JSON(http.StatusInternalServerError,
			newErrorResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)))
	}
}
