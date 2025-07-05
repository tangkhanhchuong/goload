package cache

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errCacheMiss = status.Error(codes.Internal, "failed to set data into cache")
)

type inMemoryClient struct {
	cache      map[string]any
	cacheMutex *sync.Mutex
	logger     *zap.Logger
}

func NewInMemoryClient(
	logger *zap.Logger,
) Client {
	return &inMemoryClient{
		cache:      make(map[string]any),
		cacheMutex: new(sync.Mutex),
		logger:     logger,
	}
}

// Get implements Client.
func (i *inMemoryClient) Get(ctx context.Context, key string) (any, error) {
	data, ok := i.cache[key]
	if !ok {
		return nil, errCacheMiss
	}

	return data, nil
}

// Set implements Client.
func (i *inMemoryClient) Set(ctx context.Context, key string, val any, _ time.Duration) error {
	i.cache[key] = val
	return nil
}

// AddToSet implements Client.
func (i *inMemoryClient) AddToSet(ctx context.Context, key string, val ...any) error {
	i.cacheMutex.Lock()
	defer i.cacheMutex.Unlock()

	set := i.getSet(key)
	set = append(set, val...)
	i.cache[key] = set

	return nil
}

// IsDataInSet implements Client.
func (i *inMemoryClient) IsDataInSet(ctx context.Context, key string, val any) (bool, error) {
	i.cacheMutex.Lock()
	defer i.cacheMutex.Unlock()

	set := i.getSet(key)
	for i := range set {
		if set[i] == val {
			return true, nil
		}
	}

	return false, nil
}

func (c inMemoryClient) getSet(key string) []any {
	setValue, ok := c.cache[key]
	if !ok {
		return make([]any, 0)
	}

	set, ok := setValue.([]any)
	if !ok {
		return make([]any, 0)
	}

	return set
}
