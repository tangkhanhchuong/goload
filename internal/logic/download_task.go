package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"goload/internal/dataaccess/database"
	"goload/internal/dataaccess/file"
	"goload/internal/dataaccess/mq/producer"
	"goload/internal/generated/grpc/goload"
	"goload/internal/utils"
)

const (
	downloadTaskMetadataFieldNameFileName = "file-name"
)

var (
	errNotAllowToUpdateDownloadTask = status.Error(codes.PermissionDenied, "only owners can update their download tasks")
	errNotAllowToDeleteDownloadTask = status.Error(codes.PermissionDenied, "only owners can delete their download tasks")
)

type CreateDownloadTaskInput struct {
	OfAccountID uint64
	URL         string
}

type CreateDownloadTaskOutput struct {
	DownloadTask *goload.DownloadTask
}

type GetDownloadTaskListInput struct {
	OfAccountID uint64
	Offset      uint64
	Limit       uint64
}

type GetDownloadTaskListOutput struct {
	DownloadTaskList       []*goload.DownloadTask
	TotalDownloadTaskCount uint64
}

type UpdateDownloadTaskInput struct {
	OfAccountID        uint64
	DownloadTaskID     uint64
	URL                string
	DownloadTaskStatus goload.DownloadStatus
}

type UpdateDownloadTaskOutput struct {
	Updated bool
}

type DeleteDownloadTaskInput struct {
	OfAccountID    uint64
	DownloadTaskID uint64
}

type DeleteDownloadTaskOutput struct {
	Deleted bool
}

type DownloadTaskService interface {
	UpdateDownloadTask(ctx context.Context, input UpdateDownloadTaskInput) (UpdateDownloadTaskOutput, error)
	CreateDownloadTask(ctx context.Context, input CreateDownloadTaskInput) (CreateDownloadTaskOutput, error)
	DeleteDownloadTask(ctx context.Context, input DeleteDownloadTaskInput) (DeleteDownloadTaskOutput, error)
	GetDownloadTaskList(ctx context.Context, input GetDownloadTaskListInput) (GetDownloadTaskListOutput, error)
	ExecuteDownloadTask(ctx context.Context, id uint64) error
}

type downloadTaskService struct {
	database                    *goqu.Database
	downloadTaskRepository      database.DownloadTaskRepository
	accountRepository           database.AccountRepository
	downloadTaskCreatedProvider producer.DownloadTaskCreatedProducer
	fileClient                  file.Client
	logger                      *zap.Logger
}

func NewDownloadTaskService(
	database *goqu.Database,
	downloadTaskRepository database.DownloadTaskRepository,
	accountRepository database.AccountRepository,
	downloadTaskCreatedProvider producer.DownloadTaskCreatedProducer,
	fileClient file.Client,
	logger *zap.Logger,
) DownloadTaskService {
	return &downloadTaskService{
		database:                    database,
		downloadTaskRepository:      downloadTaskRepository,
		accountRepository:           accountRepository,
		downloadTaskCreatedProvider: downloadTaskCreatedProvider,
		fileClient:                  fileClient,
		logger:                      logger,
	}
}

// CreateDownloadTask implements DownloadTaskService.
func (d *downloadTaskService) CreateDownloadTask(ctx context.Context, input CreateDownloadTaskInput) (CreateDownloadTaskOutput, error) {
	account, getAccountErr := d.accountRepository.GetAccountByID(ctx, input.OfAccountID)
	if getAccountErr != nil {
		return CreateDownloadTaskOutput{}, getAccountErr
	}

	downloadTask := database.DownloadTask{
		OfAccountID:    account.ID,
		DownloadType:   goload.DownloadType_HTTP,
		URL:            input.URL,
		DownloadStatus: goload.DownloadStatus_Pending,
		Metadata:       "{}",
	}
	txnErr := d.database.WithTx(func(td *goqu.TxDatabase) error {
		downloadTaskID, createDownloadTaskErr := d.downloadTaskRepository.
			WithDatabase(td).
			CreateDownloadTask(ctx, downloadTask)
		if createDownloadTaskErr != nil {
			return createDownloadTaskErr
		}
		downloadTask.ID = downloadTaskID

		downloadTaskCreatedEvent := producer.DownloadTaskCreatedEvent{
			DownloadTaskID: downloadTaskID,
		}
		produceErr := d.downloadTaskCreatedProvider.Produce(ctx, downloadTaskCreatedEvent)
		if produceErr != nil {
			return produceErr
		}

		return nil
	})

	if txnErr != nil {
		return CreateDownloadTaskOutput{}, txnErr
	}

	return CreateDownloadTaskOutput{
		DownloadTask: d.toProtoDownloadTask(downloadTask, account),
	}, nil
}

// DeleteDownloadTask implements DownloadTaskService.
func (d *downloadTaskService) DeleteDownloadTask(ctx context.Context, input DeleteDownloadTaskInput) (DeleteDownloadTaskOutput, error) {
	account, err := d.accountRepository.GetAccountByID(ctx, input.OfAccountID)
	if err != nil {
		return DeleteDownloadTaskOutput{}, err
	}

	downloadTask, err := d.downloadTaskRepository.GetDownloadTaskByID(ctx, input.DownloadTaskID)
	if err != nil {
		return DeleteDownloadTaskOutput{}, err
	}

	if account.ID != downloadTask.OfAccountID {
		return DeleteDownloadTaskOutput{}, errNotAllowToDeleteDownloadTask
	}

	_, err = d.downloadTaskRepository.DeleteDownloadTask(ctx, input.DownloadTaskID)
	if err != nil {
		return DeleteDownloadTaskOutput{}, err
	}

	return DeleteDownloadTaskOutput{
		Deleted: true,
	}, nil
}

// GetDownloadTaskList implements DownloadTaskService.
func (d *downloadTaskService) GetDownloadTaskList(ctx context.Context, input GetDownloadTaskListInput) (GetDownloadTaskListOutput, error) {
	account, err := d.accountRepository.GetAccountByID(ctx, input.OfAccountID)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	totalDownloadTaskCount, err := d.downloadTaskRepository.CountDownloadTasksByOfAccountID(ctx, account.ID)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	downloadTaskList, err := d.downloadTaskRepository.
		GetDownloadTaskListByOfAccountID(ctx, account.ID, input.Offset, input.Limit)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	return GetDownloadTaskListOutput{
		TotalDownloadTaskCount: totalDownloadTaskCount,
		DownloadTaskList: lo.Map(downloadTaskList, func(item database.DownloadTask, _ int) *goload.DownloadTask {
			return d.toProtoDownloadTask(item, account)
		}),
	}, nil
}

// UpdateDownloadTask implements DownloadTaskService.
func (d *downloadTaskService) UpdateDownloadTask(ctx context.Context, input UpdateDownloadTaskInput) (UpdateDownloadTaskOutput, error) {
	account, err := d.accountRepository.GetAccountByID(ctx, input.OfAccountID)
	if err != nil {
		return UpdateDownloadTaskOutput{}, err
	}

	downloadTask, err := d.downloadTaskRepository.GetDownloadTaskByID(ctx, input.DownloadTaskID)
	if err != nil {
		return UpdateDownloadTaskOutput{}, err
	}

	if account.ID != downloadTask.OfAccountID {
		return UpdateDownloadTaskOutput{}, errNotAllowToUpdateDownloadTask
	}

	if input.URL != "" {
		downloadTask.URL = input.URL
	}
	if input.DownloadTaskStatus != goload.DownloadStatus_UndefinedStatus {
		downloadTask.DownloadStatus = input.DownloadTaskStatus
	}
	_, err = d.downloadTaskRepository.UpdateDownloadTask(ctx, downloadTask)
	if err != nil {
		return UpdateDownloadTaskOutput{}, err
	}

	return UpdateDownloadTaskOutput{
		Updated: true,
	}, nil
}

// ExecuteDownloadTask implements DownloadTaskService.
func (d *downloadTaskService) ExecuteDownloadTask(ctx context.Context, id uint64) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))

	updated, downloadTask, err := d.updateDownloadTaskStatusFromPendingToDownloading(ctx, id)
	if err != nil {
		return err
	}
	if !updated {
		return nil
	}

	fileName := fmt.Sprintf("download_file_%d", id)
	fileWriterCloser, err := d.fileClient.Write(ctx, fileName)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get download file writer")
		d.updateDownloadTaskStatusToFailed(ctx, downloadTask)
		return err
	}
	defer fileWriterCloser.Close()

	var downloader Downloader
	switch downloadTask.DownloadType {
	case goload.DownloadType_HTTP:
		downloader = NewHttpDownloader(downloadTask.URL, d.logger)
	default:
		logger.With(zap.Any("download_type", downloadTask.DownloadType)).Error("unsupported download type")
		d.updateDownloadTaskStatusToFailed(ctx, downloadTask)
		return nil
	}

	metadata, err := downloader.Download(ctx, fileWriterCloser)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get download file")
		d.updateDownloadTaskStatusToFailed(ctx, downloadTask)
		return err
	}

	metadata[downloadTaskMetadataFieldNameFileName] = fileName
	downloadTask.DownloadStatus = goload.DownloadStatus_Success
	encodedMetadata, err := json.Marshal(metadata)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to stringify metadata")
		return err
	}
	downloadTask.Metadata = string(encodedMetadata)
	_, err = d.downloadTaskRepository.UpdateDownloadTask(ctx, downloadTask)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update download task status to success")
		return err
	}

	logger.With(zap.Uint64("id", id)).Info("download task is executed successfully")

	return nil
}

func (d downloadTaskService) toProtoDownloadTask(
	downloadTask database.DownloadTask,
	account database.Account,
) *goload.DownloadTask {
	return &goload.DownloadTask{
		Id: downloadTask.ID,
		OfAccount: &goload.Account{
			Id:          account.ID,
			AccountName: account.AccountName,
		},
		DownloadType:   downloadTask.DownloadType,
		Url:            downloadTask.URL,
		DownloadStatus: downloadTask.DownloadStatus,
	}
}

func (d downloadTaskService) updateDownloadTaskStatusFromPendingToDownloading(
	ctx context.Context,
	id uint64,
) (bool, database.DownloadTask, error) {
	var (
		logger       = utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))
		updated      = false
		downloadTask database.DownloadTask
		err          error
	)

	txnErr := d.database.WithTx(func(td *goqu.TxDatabase) error {
		downloadTask, err = d.downloadTaskRepository.WithDatabase(td).GetDownloadTaskByIDWithXLock(ctx, id)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to get download task for update")
			return err
		}

		if downloadTask.DownloadStatus != goload.DownloadStatus_Pending {
			logger.With(zap.Error(err)).Warn("download task is not ready to be executed")
			return nil
		}

		downloadTask.DownloadStatus = goload.DownloadStatus_Downloading
		_, err = d.downloadTaskRepository.WithDatabase(td).UpdateDownloadTask(ctx, downloadTask)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to update download task to downloading")
			return err
		}
		updated = true

		return nil
	})
	if txnErr != nil {
		return false, database.DownloadTask{}, txnErr
	}

	return updated, downloadTask, nil
}

func (d downloadTaskService) updateDownloadTaskStatusToFailed(ctx context.Context, downloadTask database.DownloadTask) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", downloadTask.ID))

	downloadTask.DownloadStatus = goload.DownloadStatus_Failed
	_, updateDownloadTaskErr := d.downloadTaskRepository.UpdateDownloadTask(ctx, downloadTask)
	if updateDownloadTaskErr != nil {
		logger.With(zap.Error(updateDownloadTaskErr)).Warn("failed to update download task status to failed")
	}
}
