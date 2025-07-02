package logic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"goload/internal/configs"
	"goload/internal/dataaccess/database"
)

const (
	rs512Bits = 2048
)

var (
	errUnexpectedSigningMethod = status.Error(codes.Unauthenticated, "unexpected signing method")
	errGetTokensClaimsFailed   = status.Error(codes.Unauthenticated, "failed to get token's claims")
	errGetTokensKidClaimFailed = status.Error(codes.Unauthenticated, "failed to get token's kid claim")
	errGetTokensSubClaimFailed = status.Error(codes.Unauthenticated, "failed to get token's sub claim")
	errGetTokensExpClaimFailed = status.Error(codes.Unauthenticated, "failed to get token's exp claim")
	errTokenPublicKeyNotFound  = status.Error(codes.Unauthenticated, "token public key not found")
	errTokenInvalidToken       = status.Error(codes.Unauthenticated, "invalid token")
	errTokenSignTokenFailed    = status.Error(codes.Internal, "failed to sign token")
)

type TokenService interface {
	GetToken(ctx context.Context, accountID uint64) (string, error)
	ParseAccountIDAndExpireTime(ctx context.Context, token string) (uint64, time.Time, error)
}

type tokenService struct {
	accountRepository   database.AccountRepository
	publicKeyRepository database.PublicKeyRepository
	authConfig          configs.Auth
	publicKeyID         uint64
	privateKey          *rsa.PrivateKey
	expiresIn           time.Duration
}

func generateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func encodePublicKeyToPEM(publicKey *rsa.PublicKey) (string, error) {
	pubASN1, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	}
	pubPEM := pem.EncodeToMemory(block)

	return string(pubPEM), nil
}

func (t tokenService) getJWTPublicKey(ctx context.Context, id uint64) (*rsa.PublicKey, error) {
	tokenPublicKey, err := t.publicKeyRepository.GetPublicKeyByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errTokenPublicKeyNotFound
		}

		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM([]byte(tokenPublicKey.PublicKey))
}

func NewTokenService(
	accounRepository database.AccountRepository,
	publicKeyRepository database.PublicKeyRepository,
	authConfig configs.Auth,
) (TokenService, error) {
	expiresIn, err := authConfig.Token.GetExpiresInDuration()
	if err != nil {
		return nil, err
	}

	privateKey, err := generateRSAKeyPair(rs512Bits)
	if err != nil {
		return nil, err
	}

	publicKeyPEM, err := encodePublicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	publicKeyID, err := publicKeyRepository.CreatePublicKey(
		context.Background(),
		database.PublicKey{PublicKey: publicKeyPEM},
	)
	if err != nil {
		return nil, err
	}

	return &tokenService{
		accountRepository:   accounRepository,
		publicKeyRepository: publicKeyRepository,
		authConfig:          authConfig,
		publicKeyID:         publicKeyID,
		privateKey:          privateKey,
		expiresIn:           expiresIn,
	}, nil
}

// GetToken implements TokenService.
func (t *tokenService) GetToken(ctx context.Context, accountID uint64) (string, error) {
	expireTime := time.Now().Add(t.expiresIn)
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"sub": accountID,
		"exp": expireTime.Unix(),
		"kid": t.publicKeyID,
	})

	tokenStr, err := token.SignedString(t.privateKey)
	if err != nil {
		return "", errTokenSignTokenFailed
	}

	return tokenStr, nil
}

// ParseAccountIDAndExpireTime implements TokenService.
func (t *tokenService) ParseAccountIDAndExpireTime(ctx context.Context, token string) (uint64, time.Time, error) {
	parsedToken, err := jwt.Parse(token, func(parsedToken *jwt.Token) (interface{}, error) {
		if _, ok := parsedToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errUnexpectedSigningMethod
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errGetTokensClaimsFailed
		}

		tokenPublicKeyID, ok := claims["kid"].(float64)
		if !ok {
			return nil, errGetTokensKidClaimFailed
		}

		return t.getJWTPublicKey(ctx, uint64(tokenPublicKeyID))
	})

	if err != nil {
		return 0, time.Time{}, errTokenInvalidToken
	}

	if !parsedToken.Valid {
		return 0, time.Time{}, errTokenInvalidToken
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, time.Time{}, errGetTokensClaimsFailed
	}

	accountID, ok := claims["sub"].(float64)
	if !ok {
		return 0, time.Time{}, errGetTokensSubClaimFailed
	}

	expireTimeUnix, ok := claims["exp"].(float64)
	if !ok {
		return 0, time.Time{}, errGetTokensExpClaimFailed
	}

	return uint64(accountID), time.Unix(int64(expireTimeUnix), 0), nil
}
