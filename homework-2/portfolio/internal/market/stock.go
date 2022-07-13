package market

import (
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/ex"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
)

type InfoService interface {
	GetStockInfo(string) (*model.StockInfo, *ex.AppError)
	GetMarketInfo() (map[string]model.StockInfo, *ex.AppError)
}
