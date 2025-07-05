package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"goload/internal/configs"
	"goload/internal/utils"
)

var (
	errSetCacheDataFailed   = status.Error(codes.Internal, "failed to set data into cache")
	errGetCacheDataFailed   = status.Error(codes.Internal, "failed to get data into cache")
	errAddDataToSetFailed   = status.Error(codes.Internal, "failed to add data into cache's set")
	errCheckCacheDataFailed = status.Error(codes.Internal, "failed to check if data in cache or not")
)

type redisClient struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisClient(
	cacheConfig configs.Cache,
	logger *zap.Logger,
) Client {
	return &redisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     cacheConfig.Address,
			Username: cacheConfig.Username,
			Password: cacheConfig.Password,
		}),
		logger: logger,
	}
}

// Get implements Client.
func (r *redisClient) Get(ctx context.Context, key string) (any, error) {
	logger := utils.LoggerWithContext(ctx, r.logger).
		With(zap.String("key", key))

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get data from cache")
		return nil, errGetCacheDataFailed
	}

	return val, nil
}

// Set implements Client.
func (r *redisClient) Set(ctx context.Context, key string, val any, duration time.Duration) error {
	logger := utils.LoggerWithContext(ctx, r.logger).
		With(zap.String("key", key)).
		With(zap.Any("val", val)).
		With(zap.Duration("duration", duration))

	if err := r.client.Set(ctx, key, val, duration).Err(); err != nil {
		logger.With(zap.Error(err)).Error("failed to set data into cache")
		return errSetCacheDataFailed
	}

	return nil
}

// AddToSet implements Client.
func (r *redisClient) AddToSet(ctx context.Context, key string, vals ...any) error {
	logger := utils.LoggerWithContext(ctx, r.logger).
		With(zap.String("key", key)).
		With(zap.Any("vals", vals))

	if err := r.client.SAdd(ctx, key, vals).Err(); err != nil {
		logger.With(zap.Error(err)).Error("failed to add data into cache's set")
		return errAddDataToSetFailed
	}

	return nil
}

// IsDataInSet implements Client.
func (r *redisClient) IsDataInSet(ctx context.Context, key string, val any) (bool, error) {
	logger := utils.LoggerWithContext(ctx, r.logger).
		With(zap.String("key", key)).
		With(zap.Any("val", val))

	exist, err := r.client.SIsMember(ctx, key, val).Result()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to check if data in cache or not")
		return false, errCheckCacheDataFailed
	}

	return exist, nil
}
