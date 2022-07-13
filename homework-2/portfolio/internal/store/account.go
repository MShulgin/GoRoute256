package store

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/ex"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/logging"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
)

type AccountRepo interface {
	GetAccountById(int32) (*model.Account, *ex.AppError)
	CreateAccount(model.Account) (*model.Account, *ex.AppError)
	FindAccount(model.Messenger, string) (*model.Account, *ex.AppError)
}

type accountRepoDB struct {
	db *sqlx.DB
}

func NewAccountsRepo(db *sqlx.DB) AccountRepo {
	return accountRepoDB{db: db}
}

func (r accountRepoDB) GetAccountById(id int32) (*model.Account, *ex.AppError) {
	acc := model.Account{}
	err := r.db.Get(&acc, "SELECT id, messenger, messenger_id FROM account WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ex.NewNotFoundError(fmt.Sprintf("Not found account: id = %d", id))
		} else {
			logging.Error("Error getting account: " + err.Error())
			return nil, ex.NewUnexpectedError("Unexpected database error")
		}
	}

	return &acc, nil
}

func (r accountRepoDB) CreateAccount(acc model.Account) (*model.Account, *ex.AppError) {
	saveSql := "INSERT INTO account (messenger, messenger_id) VALUES ($1, $2) RETURNING id"
	var accountId int32
	err := r.db.Get(&accountId, saveSql, acc.Messenger.String(), acc.MessengerId)
	if err != nil {
		logging.Error("Error while creating new account: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error from database")
	}
	acc.Id = accountId
	return &acc, nil
}

func (r accountRepoDB) FindAccount(messenger model.Messenger, messengerId string) (*model.Account, *ex.AppError) {
	acc := model.Account{}
	err := r.db.Get(&acc, "SELECT id, messenger, messenger_id FROM account WHERE messenger = $1 AND messenger_id = $2", messenger.String(), messengerId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ex.NewNotFoundError(fmt.Sprintf("Not found account: messenger = %s, messengerId = %s", messenger.String(), messengerId))
		} else {
			logging.Error("Error getting account: " + err.Error())
			return nil, ex.NewUnexpectedError("Unexpected database error")
		}
	}

	return &acc, nil
}
