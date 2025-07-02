package database

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAccountNotFound = status.Error(codes.NotFound, "account not found")
)

const (
	TabNameAccounts            = "accounts"
	ColNameAccountsID          = "id"
	ColNameAccountsAccountName = "account_name"
)

type Account struct {
	ID          uint64 `db:"id"`
	AccountName string `db:"account_name"`
}

type AccountRepository interface {
	CreateAccount(ctx context.Context, account Account) (uint64, error)
	GetAccountByID(ctx context.Context, id uint64) (Account, error)
	GetAccountByAccountName(ctx context.Context, accountName string) (Account, error)
	WithDatabase(database Database) AccountRepository
}

type accountRepository struct {
	database Database
}

func NewAccountRepository(
	database *goqu.Database,
) AccountRepository {
	return &accountRepository{
		database: database,
	}
}

// CreateAccount implements AccountRepository.
func (a *accountRepository) CreateAccount(ctx context.Context, account Account) (uint64, error) {
	var id uint64

	_, err := a.database.
		Insert(TabNameAccounts).
		Rows(goqu.Record{
			ColNameAccountsAccountName: account.AccountName,
		}).
		Returning("id").
		Executor().
		ScanValContext(ctx, &id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetAccountByAccountName implements AccountRepository.
func (a *accountRepository) GetAccountByAccountName(ctx context.Context, accountName string) (Account, error) {
	account := Account{}

	found, err := a.database.
		From(TabNameAccounts).
		Where(goqu.C(ColNameAccountsAccountName).Eq(accountName)).
		ScanStructContext(ctx, &account)
	if err != nil {
		return Account{}, err
	}
	if !found {
		return Account{}, ErrAccountNotFound
	}

	return account, nil
}

// GetAccountByID implements AccountRepository.
func (a *accountRepository) GetAccountByID(ctx context.Context, id uint64) (Account, error) {
	account := Account{}

	found, err := a.database.
		From(TabNameAccounts).
		Where(goqu.C(ColNameAccountsID).Eq(id)).
		ScanStructContext(ctx, &account)
	if err != nil {
		return Account{}, err
	}
	if !found {
		return Account{}, ErrAccountNotFound
	}

	return account, nil
}

// WithDatabase implements AccountRepository.
func (a *accountRepository) WithDatabase(database Database) AccountRepository {
	return &accountRepository{
		database: database,
	}
}
