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

type PortfolioRepo interface {
	FindByAccount(int32) ([]*model.Portfolio, *ex.AppError)
	ExistName(int32, string) (bool, *ex.AppError)
	SavePortfolio(model.Portfolio) (*model.Portfolio, *ex.AppError)
	GetById(int32) (*model.Portfolio, *ex.AppError)
	GetAll() ([]*model.Portfolio, *ex.AppError)
}

type portfolioRepoDB struct {
	db *sqlx.DB
}

func NewPortfolioRepo(db *sqlx.DB) PortfolioRepo {
	return portfolioRepoDB{db: db}
}

func (r portfolioRepoDB) FindByAccount(accountId int32) ([]*model.Portfolio, *ex.AppError) {
	var portfolioList []*model.Portfolio
	err := r.db.Select(&portfolioList, "SELECT id, label, account_id FROM portfolio WHERE account_id = $1", accountId)
	if err != nil {
		logging.Error("Error getting portfolio: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}

	if len(portfolioList) > 0 {
		idList := make([]int32, 0, len(portfolioList))
		for _, p := range portfolioList {
			idList = append(idList, p.Id)
		}
		positionSql, args, err := sqlx.In(`
				SELECT p.portfolio_id, p.id, a.code, p.quantity, p.placement_time FROM portfolio_position p
				JOIN asset a ON a.id = p.asset_id
				WHERE p.portfolio_id IN (?)
				ORDER BY p.placement_time`, idList)
		if err != nil {
			logging.Error("Failed to bind arguments to query: " + err.Error())
			return nil, ex.NewUnexpectedError("Unexpected database error")
		}
		positionSql = r.db.Rebind(positionSql)
		rows, err := r.db.Queryx(positionSql, args...)
		if err != nil {
			logging.Error("Error getting portfolio positions: " + err.Error())
			return nil, ex.NewUnexpectedError("Unexpected database error")
		}
		defer rows.Close()
		positions := make(map[int32][]model.Position, 0)
		for _, portfolioId := range idList {
			positions[portfolioId] = make([]model.Position, 0)
		}
		for rows.Next() {
			var portfolioId int32
			p := model.Position{}
			err = rows.Scan(&portfolioId, &p.Id, &p.Symbol, &p.Quantity, &p.PlacementTime)
			if err != nil {
				logging.Error("Error getting portfolio position: " + err.Error())
				return nil, ex.NewUnexpectedError("Unexpected database error")
			}
			positions[portfolioId] = append(positions[portfolioId], p)
		}
		for _, p := range portfolioList {
			p.Positions = positions[p.Id]
		}
	} else {
		for _, p := range portfolioList {
			p.Positions = make([]model.Position, 0)
		}
	}

	return portfolioList, nil
}

func (r portfolioRepoDB) SavePortfolio(portfolio model.Portfolio) (*model.Portfolio, *ex.AppError) {
	portfolioId := portfolio.Id
	if portfolioId != 0 {
		return r.updatePortfolio(portfolio)
	}

	tx, err := r.db.Beginx()
	if err != nil {
		logging.Error("Error while opening new transaction: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error from database")
	}
	defer tx.Rollback()
	portfolioSql := "INSERT INTO portfolio (label, account_id) VALUES ($1, $2) RETURNING id"
	err = tx.Get(&portfolioId, portfolioSql, portfolio.Name, portfolio.AccountId)
	if err != nil {
		logging.Error("Error while creating new portfolio: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error from database")
	}
	portfolio.Id = portfolioId
	positionSql := `INSERT INTO portfolio_position (portfolio_id, asset_id, quantity, placement_time)
					SELECT $1, a.id, $2, $3 FROM asset a WHERE code = $4
					RETURNING id`
	stmt, err := tx.Preparex(positionSql)
	if err != nil {
		logging.Error("Failed to create statement: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error from database")
	}
	defer stmt.Close()
	for _, pos := range portfolio.Positions {
		var positionId int32
		err := stmt.Get(&positionId, portfolioId, pos.Quantity, pos.PlacementTime, pos.Symbol)
		if err != nil {
			logging.Error("Error while creating new portfolio position: " + err.Error())
			return nil, ex.NewUnexpectedError("Unexpected error from database")
		}
		pos.Id = positionId
	}
	err = tx.Commit()
	if err != nil {
		logging.Error("Failed to commit transaction: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error from database")
	}
	return &portfolio, nil
}

func (r portfolioRepoDB) updatePortfolio(p model.Portfolio) (*model.Portfolio, *ex.AppError) {
	tx, err := r.db.Beginx()
	if err != nil {
		logging.Error("Error while opening new transaction: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error from database")
	}
	defer tx.Rollback()
	portfolioSql := "UPDATE portfolio SET label = $1, account_id = $2 WHERE id = $3"
	_, err = tx.Exec(portfolioSql, p.Name, p.AccountId, p.Id)
	if err != nil {
		logging.Error("Failed to update portfolio: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error from database")
	}
	dropArgs := make([]int32, 0)
	for _, pp := range p.Positions {
		if pp.Id != 0 {
			dropArgs = append(dropArgs, pp.Id)
		}
	}
	if len(dropArgs) > 0 {
		dropPositionSql, args, err := sqlx.In(
			fmt.Sprintf("DELETE FROM portfolio_position WHERE portfolio_id = %d AND id NOT IN (?)", p.Id), dropArgs)
		if err != nil {
			logging.Error("Failed to bind arguments to query: " + err.Error())
			return nil, ex.NewUnexpectedError("Unexpected database error")
		}
		dropPositionSql = r.db.Rebind(dropPositionSql)
		_, err = tx.Exec(dropPositionSql, args...)
		if err != nil {
			logging.Error("Failed to drop portfolio positions: " + err.Error())
			return nil, ex.NewUnexpectedError("Unexpected database error")
		}
	}

	positionSql := `INSERT INTO portfolio_position (portfolio_id, asset_id, quantity, placement_time)
					SELECT $1, a.id, $2, $3 FROM asset a WHERE code = $4
					RETURNING id`
	stmt, err := tx.Preparex(positionSql)
	if err != nil {
		logging.Error("Failed to create statement: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	defer stmt.Close()

	for _, pos := range p.Positions {
		if pos.Id == 0 {
			var positionId int32
			err := stmt.Get(&positionId, p.Id, pos.Quantity, pos.PlacementTime, pos.Symbol)
			if err != nil {
				logging.Error("Error while creating new portfolio position: " + err.Error())
				return nil, ex.NewUnexpectedError("Unexpected error from database")
			}
			pos.Id = positionId
		}
	}

	err = tx.Commit()
	if err != nil {
		logging.Error("Failed to commit transaction: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error from database")
	}

	return &p, nil
}

func (r portfolioRepoDB) GetById(id int32) (*model.Portfolio, *ex.AppError) {
	portfolio := model.Portfolio{}
	err := r.db.Get(&portfolio, "SELECT id, label, account_id FROM portfolio WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ex.NewNotFoundError(fmt.Sprintf("Not found portfolio: id = %d", id))
		} else {
			logging.Error("Error getting portfolio: " + err.Error())
			return nil, ex.NewUnexpectedError("Unexpected database error")
		}
	}
	var positions []model.Position
	positionSql := `SELECT p.id, a.code, p.quantity, p.placement_time FROM portfolio_position p
					JOIN asset a ON a.id = p.asset_id
					WHERE p.portfolio_id = $1
					ORDER BY p.placement_time`
	err = r.db.Select(&positions, positionSql, id)
	if err != nil {
		logging.Error("Error getting portfolio positions: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	portfolio.Positions = positions

	return &portfolio, nil
}

func (r portfolioRepoDB) GetAll() ([]*model.Portfolio, *ex.AppError) {
	portfolioList := make([]*model.Portfolio, 0)
	err := r.db.Select(&portfolioList, "SELECT id, label, account_id FROM portfolio")
	if err != nil {
		logging.Error("Error getting portfolio list: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	if len(portfolioList) == 0 {
		return make([]*model.Portfolio, 0), nil
	}
	positionSql := `SELECT p.portfolio_id, p.id, a.code, p.quantity, p.placement_time FROM portfolio_position p
					JOIN asset a ON a.id = p.asset_id`
	rows, err := r.db.Queryx(positionSql)
	if err != nil {
		logging.Error("Error getting portfolio positions: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	defer rows.Close()
	positions := make(map[int32][]model.Position)
	for _, p := range portfolioList {
		positions[p.Id] = make([]model.Position, 0)
	}
	for rows.Next() {
		var portfolioId int32
		var pos model.Position
		err = rows.Scan(&portfolioId, &pos.Id, &pos.Symbol, &pos.Quantity, &pos.PlacementTime)
		if err != nil {
			logging.Error("Failed scanning portfolio position: " + err.Error())
			return nil, ex.NewUnexpectedError("Unexpected database error")
		}
		positions[portfolioId] = append(positions[portfolioId], pos)
	}

	for _, p := range portfolioList {
		p.Positions = positions[p.Id]
	}

	return portfolioList, nil
}

func (r portfolioRepoDB) ExistName(accountId int32, name string) (bool, *ex.AppError) {
	sql := "SELECT EXISTS (SELECT * FROM portfolio WHERE account_id = $1 AND label = $2)"
	var exists bool
	err := r.db.Get(&exists, sql, accountId, name)
	if err != nil {
		logging.Error("Error checking portfolio: " + err.Error())
		return false, ex.NewUnexpectedError("Unexpected database error")
	}

	return exists, nil
}
