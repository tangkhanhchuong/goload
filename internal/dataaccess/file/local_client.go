package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"goload/internal/utils"
)

var (
	errOpenFileFailed = status.Error(codes.Internal, "failed to open file")
)

type localClient struct {
	downloadDirectory string
	logger            *zap.Logger
}

func newLocalClient(
	downloadDirectory string,
	logger *zap.Logger,
) (Client, error) {
	if err := os.MkdirAll(downloadDirectory, os.ModeDir); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("failed to create download directory: %w", err)
		}
	}

	return &localClient{
		downloadDirectory: downloadDirectory,
		logger:            logger,
	}, nil
}

// Read implements Client.
func (l *localClient) Read(ctx context.Context, filePath string) (io.ReadCloser, error) {
	logger := utils.LoggerWithContext(ctx, l.logger).With(zap.String("file_path", filePath))

	absolutePath := path.Join(l.downloadDirectory, filePath)
	file, err := os.Open(absolutePath)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to open file")
		return nil, errOpenFileFailed
	}

	return newBufferedFileReader(file), nil
}

// Write implements Client.
func (l *localClient) Write(ctx context.Context, filePath string) (io.WriteCloser, error) {
	logger := utils.LoggerWithContext(ctx, l.logger).With(zap.String("file_path", filePath))

	absolutePath := path.Join(l.downloadDirectory, filePath)
	file, err := os.Create(absolutePath)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to open file")
		return nil, errOpenFileFailed
	}

	return file, nil
}
