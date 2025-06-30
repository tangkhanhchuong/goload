package database

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

var (
	ErrAccountPasswordNotFound = status.Error(codes.NotFound, "account_password not found")
)

const (
	TabNameAccountPasswords            = "account_passwords"
	ColNameAccountPasswordsOfAccountID = "of_account_id"
	ColNameAccountPasswordsHash        = "hash"
)

type AccountPassword struct {
	OfAccountID uint64 `db:"of_account_db"`
	Hash        string `db:"hash"`
}

type AccountPasswordRepository interface {
	CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) (uint64, error)
	GetAccountPasswordByOfAccountID(ctx context.Context, ofAccountID uint64) (AccountPassword, error)
	WithDatabase(database Database) AccountPasswordRepository
}

type accountPasswordRepository struct {
	database Database
}

func NewAccountPasswordRepository(
	database *goqu.Database,
) AccountPasswordRepository {
	return &accountPasswordRepository{
		database: database,
	}
}

// CreateAccountPassword implements AccountPasswordRepository.
func (a *accountPasswordRepository) CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) (uint64, error) {
	var ofAccountId uint64
	_, err := a.database.
		Insert(TabNameAccountPasswords).
		Rows(goqu.Record{
			ColNameAccountPasswordsOfAccountID: accountPassword.OfAccountID,
			ColNameAccountPasswordsHash:        accountPassword.Hash,
		}).
		Returning("of_account_id").
		Executor().
		ScanValContext(ctx, &ofAccountId)
	if err != nil {
		fmt.Println("failed to create account password", err)
		return 0, err
	}

	return ofAccountId, nil
}

// GetAccountPasswordByOfAccountID implements AccountPasswordRepository.
func (a *accountPasswordRepository) GetAccountPasswordByOfAccountID(ctx context.Context, ofAccountID uint64) (AccountPassword, error) {
	accountPassword := AccountPassword{}
	found, err := a.database.
		From(TabNameAccountPasswords).
		Where(goqu.C(ColNameAccountPasswordsOfAccountID).Eq(ofAccountID)).
		ScanStructContext(ctx, &accountPassword)
	if err != nil {
		fmt.Println("failed to get account password by of user id", err)
		return AccountPassword{}, err
	}
	if !found {
		fmt.Println("account password is not found")
		return AccountPassword{}, ErrAccountPasswordNotFound
	}

	return accountPassword, nil
}

// WithDatabase implements AccountRepository.
func (a *accountPasswordRepository) WithDatabase(database Database) AccountPasswordRepository {
	return &accountPasswordRepository{
		database: database,
	}
}
