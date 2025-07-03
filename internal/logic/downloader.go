package logic

import (
	"context"
	"goload/internal/utils"
	"io"
	"net/http"

	"go.uber.org/zap"
)

const (
	HTTPResponseHeaderContentType = "Content-Type"
	HTTPMetadataKeyContentType    = "content-type"
)

type Downloader interface {
	Download(ctx context.Context, writer io.Writer) (map[string]any, error)
}

type httpDownloader struct {
	url    string
	logger *zap.Logger
}

func NewHttpDownloader(
	url string,
	logger *zap.Logger,
) Downloader {
	return &httpDownloader{
		url:    url,
		logger: logger,
	}
}

// Download implements Downloader.
func (h *httpDownloader) Download(ctx context.Context, writer io.Writer) (map[string]any, error) {
	logger := utils.LoggerWithContext(ctx, h.logger)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, http.NoBody)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create new http request")
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to download from url")
		return nil, err
	}
	defer response.Body.Close()

	_, err = io.Copy(writer, response.Body)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to write downloaded file")
		return nil, err
	}
	metadata := map[string]any{
		HTTPMetadataKeyContentType: response.Header.Get(HTTPResponseHeaderContentType),
	}

	return metadata, nil
}
