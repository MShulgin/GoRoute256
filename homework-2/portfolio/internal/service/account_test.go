package service

import (
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/ex"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/store"
	"testing"
)

func accounts() map[int32]*model.Account {
	return map[int32]*model.Account{
		1: {Id: 1, Messenger: model.Telegram, MessengerId: "1"},
		2: {Id: 2, Messenger: model.Telegram, MessengerId: "2"},
		3: {Id: 3, Messenger: model.Telegram, MessengerId: "3"},
	}
}

func TestGetAccountById(t *testing.T) {
	repo := store.MockAccountRepo{Values: accounts()}
	srv := NewAccountService(&repo)
	acc, err := srv.GetAccountById(2)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	expected := model.Account{Id: 2, Messenger: "Telegram", MessengerId: "2"}
	if *acc != expected {
		t.Errorf("Account %v not equal to expected %v", *acc, expected)
	}
}

func TestGetAccountByIdNotFound(t *testing.T) {
	repo := store.MockAccountRepo{Values: accounts()}
	srv := NewAccountService(&repo)
	acc, err := srv.GetAccountById(10)
	if acc != nil {
		t.Errorf("Expeted nil account value")
	}
	if err == nil || err.Code != 404 {
		t.Errorf("Expected not found error")
	}

}

func TestNewAccount(t *testing.T) {
	repo := store.MockAccountRepo{Values: accounts()}
	srv := NewAccountService(&repo)

	newAcc := model.CreateAccountReq{
		Messenger:   "Telegram",
		MessengerId: "100500",
	}

	acc, err := srv.NewAccount(newAcc)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	if acc.Id <= 0 {
		t.Errorf("Expected account id value > 0")
	}
	expected := model.Account{Id: acc.Id, Messenger: "Telegram", MessengerId: "100500"}
	if *acc != expected {
		t.Errorf("Actual account %v not equal to expected %v", *acc, expected)
	}
}

func TestNewAccountConflictError(t *testing.T) {
	repo := store.MockAccountRepo{Values: accounts()}
	srv := NewAccountService(&repo)

	newAcc := model.CreateAccountReq{
		Messenger:   "Telegram",
		MessengerId: "2",
	}

	_, err := srv.NewAccount(newAcc)

	if err == nil || !err.IsType(ex.ConflictError) {
		t.Errorf("Expected confilct error, got: %v", err)
	}
}

func TestGetAccount(t *testing.T) {
	repo := store.MockAccountRepo{Values: accounts()}
	srv := NewAccountService(&repo)

	acc, err := srv.GetAccount(model.Telegram, "2")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	expected := model.Account{Id: 2, Messenger: model.Telegram, MessengerId: "2"}
	if *acc != expected {
		t.Errorf("Actual account %v not equal to expected %v", *acc, expected)
	}
}

func TestGetAccountNotFound(t *testing.T) {
	repo := store.MockAccountRepo{Values: accounts()}
	srv := NewAccountService(&repo)

	_, err := srv.GetAccount(model.Telegram, "100")
	if err == nil || !err.IsType(ex.NotFoundError) {
		t.Errorf("Expected not found error")
	}
}
