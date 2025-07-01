package logic

import (
	"context"
	"errors"

	"github.com/doug-martin/goqu/v9"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"

	"goload/internal/dataaccess/database"
	"goload/internal/generated/grpc/goload"
)

var ErrAccountWrongPassword = status.Error(codes.Unauthenticated, "incorrect password")

type CreateAccountInput struct {
	AccountName string
	Password    string
}

type CreateAccountOutput struct {
	ID          uint64
	AccountName string
}

type CreateSessionInput struct {
	AccountName string
	Password    string
}

type CreateSessionOutput struct {
	Account goload.Account
	Token   string
}

type AccountService interface {
	CreateAccount(ctx context.Context, input CreateAccountInput) (CreateAccountOutput, error)
	CreateSession(ctx context.Context, input CreateSessionInput) (CreateSessionOutput, error)
}

type accountService struct {
	database                  *goqu.Database
	accountRepository         database.AccountRepository
	accountPasswordRepository database.AccountPasswordRepository
	hashService               HashService
	tokenService              TokenService
}

func NewAccountService(
	database *goqu.Database,
	accountRepository database.AccountRepository,
	accountPasswordRepository database.AccountPasswordRepository,
	hashService HashService,
	tokenService TokenService,
) AccountService {
	return &accountService{
		database:                  database,
		accountRepository:         accountRepository,
		accountPasswordRepository: accountPasswordRepository,
		hashService:               hashService,
		tokenService:              tokenService,
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

// CreateSession implements AccountService.
func (a *accountService) CreateSession(ctx context.Context, input CreateSessionInput) (CreateSessionOutput, error) {
	foundAccount, err := a.accountRepository.GetAccountByAccountName(ctx, input.AccountName)
	if err != nil {
		return CreateSessionOutput{}, err
	}

	foundAccountPassword, err := a.accountPasswordRepository.GetAccountPasswordByOfAccountID(ctx, foundAccount.ID)
	if err != nil {
		return CreateSessionOutput{}, err
	}

	isHashEqual, err := a.hashService.IsHashEqual(ctx, input.Password, foundAccountPassword.Hash)
	if err != nil {
		return CreateSessionOutput{}, err
	}
	if !isHashEqual {
		return CreateSessionOutput{}, ErrAccountWrongPassword
	}

	token, err := a.tokenService.GetToken(ctx, foundAccount.ID)
	if err != nil {
		return CreateSessionOutput{}, err
	}

	return CreateSessionOutput{
		Account: goload.Account{
			Id:          foundAccount.ID,
			AccountName: foundAccount.AccountName,
		},
		Token: token,
	}, nil
}
