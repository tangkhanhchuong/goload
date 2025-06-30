package logic

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"

	"goload/internal/configs"
)

var (
	ErrHashHashDataFailed      = status.Error(codes.Internal, "failed to hash data")
	ErrHashCompareHahsedFailed = status.Error(codes.Internal, "failed to compare hashed")
)

type HashService interface {
	Hash(ctx context.Context, data string) (string, error)
	IsHashEqual(ctx context.Context, data string, hashed string) (bool, error)
}

type hashService struct {
	authConfig configs.Auth
}

func NewHashService(authConfig configs.Auth) HashService {
	return &hashService{
		authConfig: authConfig,
	}
}

// Hash implements HashService.
func (h *hashService) Hash(ctx context.Context, data string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(data), h.authConfig.Hash.Cost)
	if err != nil {
		return "", ErrHashHashDataFailed
	}

	return string(hashed), nil
}

// IsHashEqual implements HashService.
func (h *hashService) IsHashEqual(ctx context.Context, data string, hashed string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(data))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, ErrHashCompareHahsedFailed
	}

	return true, nil
}
