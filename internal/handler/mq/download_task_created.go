package mq

import (
	"context"

	"go.uber.org/zap"

	"goload/internal/dataaccess/mq/producer"
	"goload/internal/logic"
	"goload/internal/utils"
)

type DownloadTaskCreated interface {
	Handle(ctx context.Context, event producer.DownloadTaskCreatedEvent) error
}

type downloadTaskCreated struct {
	downloadTaskService logic.DownloadTaskService
	logger              *zap.Logger
}

func NewDownloadTaskCreated(
	downloadTaskService logic.DownloadTaskService,
	logger *zap.Logger,
) DownloadTaskCreated {
	return &downloadTaskCreated{
		downloadTaskService: downloadTaskService,
		logger:              logger,
	}
}

func (d downloadTaskCreated) Handle(ctx context.Context, event producer.DownloadTaskCreatedEvent) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("event", event))
	logger.Info("download task created event received")

	if err := d.downloadTaskService.ExecuteDownloadTask(ctx, event.DownloadTaskID); err != nil {
		logger.With(zap.Error(err)).Error("failed to handle download task created event")
		return err
	}

	return nil
}
