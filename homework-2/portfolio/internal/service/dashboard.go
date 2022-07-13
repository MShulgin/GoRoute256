package service

import (
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/ex"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/store"
	"sort"
)

type DashboardService interface {
	GetAccountDashboard(accountId int32) (*model.AccountDashboard, *ex.AppError)
}

func NewDefaultDashboardService(portfolioRepo store.PortfolioRepo, valueRepo store.PortfolioValueRepo) DashboardService {
	return DefaultDashboardService{
		portfolioRepo: portfolioRepo,
		valueRepo:     valueRepo,
	}
}

type DefaultDashboardService struct {
	portfolioRepo store.PortfolioRepo
	valueRepo     store.PortfolioValueRepo
}

func (d DefaultDashboardService) GetAccountDashboard(accountId int32) (*model.AccountDashboard, *ex.AppError) {
	portfolioList, err := d.portfolioRepo.FindByAccount(accountId)
	if err != nil {
		return nil, err
	}
	portfolioValue := 0.0
	portfolioValueList := make([]model.PortfolioValueInfo, 0)
	for _, p := range portfolioList {
		valueHistory, err := d.valueRepo.GetPortfolioHistory(p.Id)
		if err != nil {
			return nil, err
		}
		portfolioValue += valueHistory.CurrentValue()
		portfolioValueInfo := model.PortfolioValueInfo{
			PortfolioId:   p.Id,
			PortfolioName: p.Name,
			Value:         valueHistory.CurrentValue(),
		}
		portfolioValueList = append(portfolioValueList, portfolioValueInfo)
	}
	sort.Slice(portfolioValueList, func(i, j int) bool {
		return portfolioValueList[i].PortfolioId < portfolioValueList[j].PortfolioId
	})
	dashboard := model.AccountDashboard{
		AccountId:          accountId,
		TotalValue:         portfolioValue,
		PortfolioValueList: portfolioValueList,
	}

	return &dashboard, nil
}
