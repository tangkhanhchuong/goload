package logic

import (
	"context"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"

	"goload/internal/dataaccess/database"
)

type CreateAccountInput struct {
	AccountName string
	Password    string
}

type CreateAccountOutput struct {
	ID          uint64
	AccountName string
}

type AccountService interface {
	CreateAccount(ctx context.Context, input CreateAccountInput) (CreateAccountOutput, error)
}

type accountService struct {
	database                  *goqu.Database
	accountRepository         database.AccountRepository
	accountPasswordRepository database.AccountPasswordRepository
	hashService               HashService
}

func NewAccountService(
	database *goqu.Database,
	accountRepository database.AccountRepository,
	accountPasswordRepository database.AccountPasswordRepository,
	hashService HashService,
) AccountService {
	return &accountService{
		database:                  database,
		accountRepository:         accountRepository,
		accountPasswordRepository: accountPasswordRepository,
		hashService:               hashService,
	}
}

func (a accountService) isAccountNameTaken(ctx context.Context, accountName string) (bool, error) {
	if _, err := a.accountRepository.GetAccountByAccountName(ctx, accountName); err != nil {
		if errors.Is(err, database.ErrAccountNotFound) {
			return false, nil
		}
		return true, err
	}
	return true, nil
}

// CreateAccount implements AccountService.
func (a *accountService) CreateAccount(ctx context.Context, input CreateAccountInput) (CreateAccountOutput, error) {
	isAccountNameTaken, err := a.isAccountNameTaken(ctx, input.AccountName)
	if err != nil {
		return CreateAccountOutput{}, err
	}
	if isAccountNameTaken {
		return CreateAccountOutput{}, errors.New("account name is already taken")
	}

	var accountId uint64
	txnErr := a.database.WithTx(func(txn *goqu.TxDatabase) error {
		accountId, err = a.accountRepository.
			WithDatabase(txn).
			CreateAccount(ctx, database.Account{
				AccountName: input.AccountName,
			})
		if err != nil {
			return err
		}

		hashedPassword, err := a.hashService.Hash(ctx, input.Password)
		if err != nil {
			fmt.Println("failed to hash password")
			return err
		}

		_, err = a.accountPasswordRepository.
			WithDatabase(txn).
			CreateAccountPassword(ctx, database.AccountPassword{
				OfAccountID: accountId,
				Hash:        hashedPassword,
			})
		if err != nil {
			return err
		}

		return nil
	})
	if txnErr != nil {
		return CreateAccountOutput{}, txnErr
	}

	return CreateAccountOutput{
		ID:          accountId,
		AccountName: input.AccountName,
	}, nil
}
