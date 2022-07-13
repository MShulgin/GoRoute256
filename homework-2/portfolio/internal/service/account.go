package service

import (
	"fmt"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/ex"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/store"
)

type AccountService interface {
	GetAccountById(int32) (*model.Account, *ex.AppError)
	GetAccount(model.Messenger, string) (*model.Account, *ex.AppError)
	NewAccount(model.CreateAccountReq) (*model.Account, *ex.AppError)
}

type DefaultAccountService struct {
	accountRepo store.AccountRepo
}

func NewAccountService(accountRepo store.AccountRepo) DefaultAccountService {
	return DefaultAccountService{accountRepo: accountRepo}
}

func (srv DefaultAccountService) GetAccountById(id int32) (*model.Account, *ex.AppError) {
	return srv.accountRepo.GetAccountById(id)
}

func (srv DefaultAccountService) NewAccount(r model.CreateAccountReq) (*model.Account, *ex.AppError) {
	_, err := srv.accountRepo.FindAccount(r.Messenger, r.MessengerId)
	if err == nil {
		errMsg := fmt.Sprintf("Account already exists: messenger = %s, messengerId = %s", r.Messenger, r.MessengerId)
		return nil, ex.NewConflictError(errMsg)
	}
	if !err.IsType(ex.NotFoundError) {
		return nil, err
	}
	acc := model.Account{Messenger: r.Messenger, MessengerId: r.MessengerId}
	return srv.accountRepo.CreateAccount(acc)
}

func (srv DefaultAccountService) GetAccount(messenger model.Messenger, messengerId string) (*model.Account, *ex.AppError) {
	return srv.accountRepo.FindAccount(messenger, messengerId)
}
