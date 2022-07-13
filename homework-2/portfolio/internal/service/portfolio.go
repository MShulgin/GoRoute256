package service

import (
	"fmt"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/ex"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/logging"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/market"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/store"
	"time"
)

type PortfolioService interface {
	GetAccountPortfolio(int32) ([]*model.Portfolio, *ex.AppError)
	GetById(int32) (*model.Portfolio, *ex.AppError)
	NewPortfolio(model.NewPortfolioRequest) (*model.Portfolio, *ex.AppError)
	PlacePosition(model.NewPositionRequest) (*model.Position, *ex.AppError)
	DeletePosition(int32, int32) *ex.AppError
	UpdatePortfolioValues() *ex.AppError
}

type DefaultPortfolioService struct {
	portfolioRepo store.PortfolioRepo
	assetRepo     store.AssetRepo
	valueRepo     store.PortfolioValueRepo
	marketClient  market.InfoService
}

func NewPortfolioService(
	portfolioRepo store.PortfolioRepo,
	assetRepo store.AssetRepo,
	valueRepo store.PortfolioValueRepo,
	marketClient market.InfoService) DefaultPortfolioService {
	return DefaultPortfolioService{
		portfolioRepo: portfolioRepo,
		assetRepo:     assetRepo,
		valueRepo:     valueRepo,
		marketClient:  marketClient,
	}
}

func (d DefaultPortfolioService) GetAccountPortfolio(accountId int32) ([]*model.Portfolio, *ex.AppError) {
	return d.portfolioRepo.FindByAccount(accountId)
}

func (d DefaultPortfolioService) GetById(id int32) (*model.Portfolio, *ex.AppError) {
	return d.portfolioRepo.GetById(id)
}

func (d DefaultPortfolioService) NewPortfolio(req model.NewPortfolioRequest) (*model.Portfolio, *ex.AppError) {
	isExist, err := d.portfolioRepo.ExistName(req.AccountId, req.Name)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, ex.NewConflictError(fmt.Sprintf("Portfolio with name '%s' already exists", req.Name))
	}
	portfolio := model.Portfolio{
		Id:        0,
		Name:      req.Name,
		AccountId: req.AccountId,
		Positions: nil,
	}
	p, err := d.portfolioRepo.SavePortfolio(portfolio)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (d DefaultPortfolioService) PlacePosition(request model.NewPositionRequest) (*model.Position, *ex.AppError) {
	portfolioId := request.PortfolioId
	if request.Quantity <= 0 {
		return nil, ex.NewBadRequestError("Quantity value must be > 0")
	}
	portfolio, err := d.portfolioRepo.GetById(portfolioId)
	if err != nil {
		return nil, err
	}
	asset := model.Asset{Code: request.Symbol}
	err = d.assetRepo.SaveAsset(asset)
	if err != nil {
		return nil, err
	}
	newPosition := model.NewPosition(request.Symbol, request.Quantity)

	portfolio.AddPosition(newPosition)
	_, err = d.portfolioRepo.SavePortfolio(*portfolio)
	if err != nil {
		return nil, err
	}

	err = d.UpdatePortfolioValues()
	if err != nil {
		logging.Error("Fail to update portfolio value: " + err.Error())
	}

	return &newPosition, nil
}

func (d DefaultPortfolioService) DeletePosition(portfolioId int32, positionId int32) *ex.AppError {
	portfolio, err := d.portfolioRepo.GetById(portfolioId)
	if err != nil {
		return err
	}
	portfolio.DeletePosition(positionId)
	_, err = d.portfolioRepo.SavePortfolio(*portfolio)
	if err != nil {
		return err
	}

	return nil
}

func (d DefaultPortfolioService) UpdatePortfolioValue(portfolioId int32) *ex.AppError {
	marketInfo, err := d.marketClient.GetMarketInfo()
	if err != nil {
		return err
	}
	portfolioList, err := d.portfolioRepo.GetById(portfolioId)
	if err != nil {
		return err
	}
	portfolioValues := calculateActualValues([]*model.Portfolio{portfolioList}, marketInfo)
	for _, v := range portfolioValues {
		err := d.valueRepo.SavePortfolioValue(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d DefaultPortfolioService) UpdatePortfolioValues() *ex.AppError {
	marketInfo, err := d.marketClient.GetMarketInfo()
	if err != nil {
		return err
	}
	portfolioList, err := d.portfolioRepo.GetAll()
	if err != nil {
		return err
	}
	portfolioValues := calculateActualValues(portfolioList, marketInfo)
	for _, v := range portfolioValues {
		err := d.valueRepo.SavePortfolioValue(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func calculateActualValues(portfolioList []*model.Portfolio, marketInfo map[string]model.StockInfo) []model.PortfolioValue {
	values := make([]model.PortfolioValue, 0, len(portfolioList))
	calcTime := time.Now()
	for _, p := range portfolioList {
		portfolioSum := 0.0
		for _, pp := range p.Positions {
			if info, ok := marketInfo[pp.Symbol]; ok {
				portfolioSum += float64(pp.Quantity) * info.LastPrice
			}
		}
		portfolioValue := model.PortfolioValue{
			PortfolioId:     p.Id,
			Value:           portfolioSum,
			CalculationTime: calcTime,
		}

		values = append(values, portfolioValue)
	}
	return values
}
