package file

import (
	"context"
	"fmt"
	"goload/internal/configs"
	"io"

	"go.uber.org/zap"
)

type Client interface {
	Write(ctx context.Context, filePath string) (io.WriteCloser, error)
	Read(ctx context.Context, filePath string) (io.ReadCloser, error)
}

func NewClient(
	downloadConfig configs.Download,
	logger *zap.Logger,
) (Client, error) {
	switch downloadConfig.Mode {
	case configs.DownloadModeLocal:
		return newLocalClient(downloadConfig.DownloadDirectory, logger)
	default:
		return nil, fmt.Errorf("download mode is unsupported: %s", downloadConfig.Mode)
	}
}
