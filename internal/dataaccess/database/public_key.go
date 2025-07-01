package database

import (
	"context"

	"github.com/doug-martin/goqu/v9"

	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

var (
	ErrPublicKeyNotFound = status.Error(codes.NotFound, "public_key not found")
)

const (
	TabNamePublicKeys          = "public_keys"
	ColNamePublicKeysID        = "id"
	ColNamePublicKeysPublicKey = "public_key"
)

type PublicKey struct {
	ID        uint64 `db:"id"`
	PublicKey string `db:"public_key"`
}

type PublicKeyRepository interface {
	CreatePublicKey(ctx context.Context, publicKey PublicKey) (uint64, error)
	GetPublicKeyByID(ctx context.Context, id uint64) (PublicKey, error)
	WithDatabase(database Database) PublicKeyRepository
}

type publicKeyRepository struct {
	database Database
}

// WithDatabase implements PublicKeyRepository.
func (p *publicKeyRepository) WithDatabase(database Database) PublicKeyRepository {
	panic("unimplemented")
}

func NewPublicKeyRepository(
	database *goqu.Database,
) PublicKeyRepository {
	return &publicKeyRepository{
		database: database,
	}
}

// CreatePublicKey implements PublicKeyRepository.
func (p *publicKeyRepository) CreatePublicKey(ctx context.Context, publicKey PublicKey) (uint64, error) {
	var publicKeyId uint64

	_, err := p.database.
		Insert(TabNamePublicKeys).
		Rows(goqu.Record{
			ColNamePublicKeysPublicKey: publicKey.PublicKey,
		}).
		Returning("id").
		Executor().
		ScanValContext(ctx, &publicKeyId)
	if err != nil {
		return 0, err
	}

	return publicKeyId, nil
}

// GetPublicKeyByID implements PublicKeyRepository.
func (p *publicKeyRepository) GetPublicKeyByID(ctx context.Context, id uint64) (PublicKey, error) {
	publicKey := PublicKey{}

	found, err := p.database.
		From(TabNamePublicKeys).
		Where(goqu.C(ColNamePublicKeysID).Eq(id)).
		ScanStructContext(ctx, &publicKey)
	if err != nil {
		return PublicKey{}, err
	}
	if !found {
		return PublicKey{}, ErrPublicKeyNotFound
	}

	return publicKey, nil
}
