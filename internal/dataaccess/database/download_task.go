package database

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"goload/internal/generated/grpc/goload"
	"goload/internal/utils"
)

var (
	errCreateDownloadTaskFailed  = status.Error(codes.Internal, "failed to create download task")
	errDeleteDownloadTaskFailed  = status.Error(codes.Internal, "failed to delete download task")
	errUpdateDownloadTaskFailed  = status.Error(codes.Internal, "failed to update download task")
	errGetDownloadTaskListFailed = status.Error(codes.Internal, "failed to get download task list of account")
	errCountDownloadTasksFailed  = status.Error(codes.Internal, "failed to count download task of account")

	ErrDownloadTaskNotFound = status.Error(codes.NotFound, "download task not found")
)

const (
	TabNameDownloadTasks               = "download_tasks"
	ColNameDownloadTasksID             = "id"
	ColNameDownloadTasksOfAccountID    = "of_account_id"
	ColNameDownloadTasksDownloadType   = "download_type"
	ColNameDownloadTasksURL            = "url"
	ColNameDownloadTasksDownloadStatus = "download_status"
	ColNameDownloadTasksMetadata       = "metadata"
)

type DownloadTask struct {
	ID             uint64                `db:"id" goqu:"skipinsert,skipupdate"`
	OfAccountID    uint64                `db:"of_account_id" goqu:"skipupdate"`
	DownloadType   goload.DownloadType   `db:"download_type"`
	URL            string                `db:"url"`
	DownloadStatus goload.DownloadStatus `db:"download_status"`
	Metadata       string                `db:"metadata"`
}

type DownloadTaskRepository interface {
	CreateDownloadTask(ctx context.Context, downloadTask DownloadTask) (uint64, error)
	UpdateDownloadTask(ctx context.Context, downloadTask DownloadTask) (bool, error)
	DeleteDownloadTask(ctx context.Context, id uint64) (bool, error)
	GetDownloadTaskListByOfAccountID(ctx context.Context, accountID, offset, limit uint64) ([]DownloadTask, error)
	CountDownloadTasksByOfAccountID(ctx context.Context, accountID uint64) (uint64, error)
	GetDownloadTaskByID(ctx context.Context, id uint64) (DownloadTask, error)
	WithDatabase(database Database) DownloadTaskRepository
}

type downloadTaskRepository struct {
	database Database
	logger   *zap.Logger
}

func NewDownloadRepository(
	database *goqu.Database,
	logger *zap.Logger,
) DownloadTaskRepository {
	return &downloadTaskRepository{
		database: database,
		logger:   logger,
	}
}

// CreateDownloadTask implements DownloadTaskRepository.
func (d *downloadTaskRepository) CreateDownloadTask(ctx context.Context, downloadTask DownloadTask) (uint64, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("task", downloadTask))
	var id uint64

	_, err := d.database.
		Insert(TabNameDownloadTasks).
		Rows(goqu.Record{
			ColNameDownloadTasksOfAccountID:    downloadTask.OfAccountID,
			ColNameDownloadTasksURL:            downloadTask.URL,
			ColNameDownloadTasksDownloadType:   downloadTask.DownloadType,
			ColNameDownloadTasksDownloadStatus: downloadTask.DownloadStatus,
			ColNameDownloadTasksMetadata:       downloadTask.Metadata,
		}).
		Returning("id").
		Executor().
		ScanValContext(ctx, &id)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create download task")
		return 0, errCreateDownloadTaskFailed
	}

	return id, nil
}

// DeleteDownloadTask implements DownloadTaskRepository.
func (d *downloadTaskRepository) DeleteDownloadTask(ctx context.Context, id uint64) (bool, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))

	if _, err := d.database.
		Delete(TabNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTasksID: id}).
		Executor().
		ExecContext(ctx); err != nil {
		logger.With(zap.Error(err)).Error("failed to delete download task")
		return false, errDeleteDownloadTaskFailed
	}

	return true, nil
}

// UpdateDownloadTask implements DownloadTaskRepository.
func (d *downloadTaskRepository) UpdateDownloadTask(ctx context.Context, downloadTask DownloadTask) (bool, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("task", downloadTask))
	if _, err := d.database.
		Update(TabNameDownloadTasks).
		Set(downloadTask).
		Where(goqu.Ex{ColNameDownloadTasksID: downloadTask.ID}).
		Executor().
		ExecContext(ctx); err != nil {
		logger.With(zap.Error(err)).Error("failed to update download task")
		return false, errUpdateDownloadTaskFailed
	}

	return true, nil
}

// CountDownloadTasksByOfAccountID implements DownloadTaskRepository.
func (d *downloadTaskRepository) CountDownloadTasksByOfAccountID(ctx context.Context, accountID uint64) (uint64, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("account_id", accountID))

	count, err := d.database.
		From(TabNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTasksOfAccountID: accountID}).
		CountContext(ctx)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to count download task of user")
		return 0, errCountDownloadTasksFailed
	}

	return uint64(count), nil
}

// GetDownloadTaskListByOfAccountID implements DownloadTaskRepository.
func (d *downloadTaskRepository) GetDownloadTaskListByOfAccountID(ctx context.Context, accountID uint64, offset uint64, limit uint64) ([]DownloadTask, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).
		With(zap.Uint64("account_id", accountID)).
		With(zap.Uint64("offset", offset)).
		With(zap.Uint64("limit", limit))

	downloadTaskList := make([]DownloadTask, 0)
	err := d.database.
		Select().
		From(TabNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTasksOfAccountID: accountID}).
		Offset(uint(offset)).
		Limit(uint(limit)).
		Executor().
		ScanStructsContext(ctx, &downloadTaskList)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get download task list of account")
		return nil, errGetDownloadTaskListFailed
	}

	return downloadTaskList, nil
}

// GetDownloadTaskByID implements DownloadTaskRepository.
func (d *downloadTaskRepository) GetDownloadTaskByID(ctx context.Context, id uint64) (DownloadTask, error) {
	downloadTask := DownloadTask{}

	found, err := d.database.
		From(TabNameDownloadTasks).
		Where(goqu.C(ColNameDownloadTasksID).Eq(id)).
		ScanStructContext(ctx, &downloadTask)
	if err != nil {
		return DownloadTask{}, err
	}
	if !found {
		return DownloadTask{}, ErrDownloadTaskNotFound
	}

	return downloadTask, nil
}

// WithDatabase implements DownloadTaskRepository.
func (d *downloadTaskRepository) WithDatabase(database Database) DownloadTaskRepository {
	return &downloadTaskRepository{
		database: database,
		logger:   d.logger,
	}
}
