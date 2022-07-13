package store

import (
	"github.com/jmoiron/sqlx"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/ex"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/logging"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
)

type PortfolioValueRepo interface {
	GetPortfolioHistory(int32) (*model.PortfolioValueHistory, *ex.AppError)
	SavePortfolioValue(value model.PortfolioValue) *ex.AppError
}

type portfolioValueRepoDb struct {
	db *sqlx.DB
}

func NewPortfolioValueRepo(db *sqlx.DB) PortfolioValueRepo {
	return portfolioValueRepoDb{db: db}
}

func (r portfolioValueRepoDb) GetPortfolioHistory(portfolioId int32) (*model.PortfolioValueHistory, *ex.AppError) {
	sql := "SELECT portfolio_id, value, calculation_time FROM portfolio_value_history WHERE portfolio_id = $1"
	var valueList []model.PortfolioValue
	err := r.db.Select(&valueList, sql, portfolioId)
	if err != nil {
		logging.Error("Error getting portfolio history: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	valueHistory := model.NewPortfolioValueHistory(valueList)
	return valueHistory, nil
}

func (r portfolioValueRepoDb) SavePortfolioValue(pv model.PortfolioValue) *ex.AppError {
	sql := "INSERT INTO portfolio_value_history(portfolio_id, value, calculation_time) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(sql, pv.PortfolioId, pv.Value, pv.CalculationTime)
	if err != nil {
		logging.Error("Failed to save portfolio value: " + err.Error())
		return ex.NewUnexpectedError("Unexpected database error")
	}
	return nil
}
