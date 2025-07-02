package producer

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"goload/internal/utils"
)

const (
	MessageQueueTopicDownloadTaskCreated = "topic-download_task_created"
)

var (
	errMarshalDownloadTaskEventFailed = status.Error(codes.Internal, "failed to marshal download task created event")
	errProduceDownloadTaskEventFailed = status.Error(codes.Internal, "failed to produce download task created event")
)

type DownloadTaskCreatedEvent struct {
	DownloadTaskID uint64 `json:"download_task_id"`
}

type DownloadTaskCreatedProducer interface {
	Produce(ctx context.Context, event DownloadTaskCreatedEvent) error
}

type downloadTaskCreatedProducer struct {
	client Client
	logger *zap.Logger
}

func NewDownloadTaskCreatedProducer(
	client Client,
	logger *zap.Logger,
) DownloadTaskCreatedProducer {
	return &downloadTaskCreatedProducer{
		client: client,
		logger: logger,
	}
}

// Produce implements DownloadTaskCreatedProducer.
func (d *downloadTaskCreatedProducer) Produce(ctx context.Context, event DownloadTaskCreatedEvent) error {
	logger := utils.LoggerWithContext(ctx, d.logger)

	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to marshal download task created event")
		return errMarshalDownloadTaskEventFailed
	}

	err = d.client.Produce(ctx, MessageQueueTopicDownloadTaskCreated, eventBytes)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to produce download task created event")
		return errProduceDownloadTaskEventFailed
	}

	return nil
}
