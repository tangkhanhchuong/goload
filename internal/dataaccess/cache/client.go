package cache

import (
	"context"
	"fmt"
	"goload/internal/configs"
	"time"

	"go.uber.org/zap"
)

type Client interface {
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Get(ctx context.Context, key string) (any, error)
	AddToSet(ctx context.Context, key string, val ...any) error
	IsDataInSet(ctx context.Context, key string, val any) (bool, error)
}

func NewClient(
	cacheConfig configs.Cache,
	logger *zap.Logger,
) (Client, error) {
	switch cacheConfig.Type {
	case configs.CacheTypeRedis:
		return NewRedisClient(cacheConfig, logger), nil
	case configs.CacheTypeInMemory:
		return NewInMemoryClient(logger), nil
	default:
		return nil, fmt.Errorf("unsupported cache type %s", cacheConfig.Type)
	}
}
