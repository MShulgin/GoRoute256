package service

import (
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/store"
	"math"
	"testing"
	"time"
)

func TestGetAccountDashboard(t *testing.T) {
	accountId := int32(1)
	portfolioMap := map[int32]*model.Portfolio{
		1: {Id: 1, Name: "Test", AccountId: accountId, Positions: []model.Position{{Id: 1, Symbol: "OZON", Quantity: 10, PlacementTime: time.Now()}}},
		2: {Id: 2, Name: "Test2", AccountId: accountId, Positions: []model.Position{{Id: 2, Symbol: "YNDX", Quantity: 50, PlacementTime: time.Now()}}},
	}

	portfolioValueHistory := []*model.PortfolioValue{
		{PortfolioId: 1, Value: 300.0, CalculationTime: time.Now()},
		{PortfolioId: 2, Value: 500.0, CalculationTime: time.Now()},
	}

	portfolioRepo := store.MockPortfolioRepo{Values: portfolioMap}
	portfolioValueRepo := store.MockPortfolioValueRepo{Values: portfolioValueHistory}
	dashboardService := NewDefaultDashboardService(&portfolioRepo, &portfolioValueRepo)

	actual, err := dashboardService.GetAccountDashboard(accountId)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}
	expected := model.AccountDashboard{
		AccountId:  accountId,
		TotalValue: 800.0,
		PortfolioValueList: []model.PortfolioValueInfo{
			{PortfolioId: 1, PortfolioName: "Test", Value: 300.0},
			{PortfolioId: 2, PortfolioName: "Test2", Value: 500.0},
		},
	}

	if actual.AccountId != expected.AccountId {
		t.Errorf("Account id %d not equal to expected %d", actual.AccountId, expected.AccountId)
	}
	if math.Abs(actual.TotalValue-expected.TotalValue) > 0.001 {
		t.Errorf("Total value of actual dashboard %.2f not equal to expected %.2f", actual.TotalValue, expected.TotalValue)
	}
	if len(actual.PortfolioValueList) != len(expected.PortfolioValueList) {
		t.Errorf("Actual len list %d not equal to expected %d", len(actual.PortfolioValueList), len(expected.PortfolioValueList))
	}
	for i := range actual.PortfolioValueList {
		if !actual.PortfolioValueList[i].Cmp(expected.PortfolioValueList[i]) {
			t.Errorf("Portfolio value %v not equal to %v", actual.PortfolioValueList[i], expected.PortfolioValueList[i])
		}
	}
}

func TestGetAccountDashboardEmpty(t *testing.T) {
	accountId := int32(1)
	portfolioMap := map[int32]*model.Portfolio{}

	var portfolioValueHistory []*model.PortfolioValue

	portfolioRepo := store.MockPortfolioRepo{Values: portfolioMap}
	portfolioValueRepo := store.MockPortfolioValueRepo{Values: portfolioValueHistory}
	dashboardService := NewDefaultDashboardService(&portfolioRepo, &portfolioValueRepo)
	actual, err := dashboardService.GetAccountDashboard(accountId)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}

	expected := model.AccountDashboard{
		AccountId:          accountId,
		TotalValue:         0.0,
		PortfolioValueList: make([]model.PortfolioValueInfo, 0),
	}
	if actual.AccountId != expected.AccountId {
		t.Errorf("Account id %d not equal to expected %d", actual.AccountId, expected.AccountId)
	}
	if math.Abs(actual.TotalValue-expected.TotalValue) > 0.001 {
		t.Errorf("Total value of actual dashboard %.2f not equal to expected %.2f", actual.TotalValue, expected.TotalValue)
	}
	if len(actual.PortfolioValueList) != len(expected.PortfolioValueList) {
		t.Errorf("Actual len list %d not equal to expected %d", len(actual.PortfolioValueList), len(expected.PortfolioValueList))
	}
	for i := range actual.PortfolioValueList {
		if !actual.PortfolioValueList[i].Cmp(expected.PortfolioValueList[i]) {
			t.Errorf("Portfolio value %v not equal to %v", actual.PortfolioValueList[i], expected.PortfolioValueList[i])
		}
	}
}
