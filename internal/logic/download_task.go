package logic

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"goload/internal/dataaccess/database"
	"goload/internal/dataaccess/mq/producer"
	"goload/internal/generated/grpc/goload"
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
	CreateDownloadTask(ctx context.Context, input CreateDownloadTaskInput) (CreateDownloadTaskOutput, error)
	UpdateDownloadTask(ctx context.Context, input UpdateDownloadTaskInput) (UpdateDownloadTaskOutput, error)
	DeleteDownloadTask(ctx context.Context, input DeleteDownloadTaskInput) (DeleteDownloadTaskOutput, error)
	GetDownloadTaskList(ctx context.Context, input GetDownloadTaskListInput) (GetDownloadTaskListOutput, error)
}

type downloadTaskService struct {
	database                    *goqu.Database
	downloadTaskRepository      database.DownloadTaskRepository
	accountRepository           database.AccountRepository
	downloadTaskCreatedProvider producer.DownloadTaskCreatedProducer
}

func NewDownloadTaskService(
	database *goqu.Database,
	downloadTaskRepository database.DownloadTaskRepository,
	accountRepository database.AccountRepository,
	downloadTaskCreatedProvider producer.DownloadTaskCreatedProducer,
) DownloadTaskService {
	return &downloadTaskService{
		database:                    database,
		downloadTaskRepository:      downloadTaskRepository,
		accountRepository:           accountRepository,
		downloadTaskCreatedProvider: downloadTaskCreatedProvider,
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
