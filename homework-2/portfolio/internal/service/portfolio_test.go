package service

import (
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/ex"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/store"
	"math"
	"testing"
	"time"
)

type mockMarketClient struct {
	Values map[string]*model.StockInfo
}

func (m mockMarketClient) GetStockInfo(s string) (*model.StockInfo, *ex.AppError) {
	if x, ok := m.Values[s]; ok {
		return x, nil
	} else {
		return nil, ex.NewNotFoundError("Not found code")
	}
}

func (m mockMarketClient) GetMarketInfo() (map[string]model.StockInfo, *ex.AppError) {
	result := make(map[string]model.StockInfo)
	for k := range m.Values {
		result[k] = *m.Values[k]
	}
	return result, nil
}

func portfolioList() map[int32]*model.Portfolio {
	return map[int32]*model.Portfolio{
		1: {Id: 1, Name: "Test", AccountId: 1, Positions: []model.Position{{Id: 1, Symbol: "OZON", Quantity: 10, PlacementTime: time.Time{}}}},
		2: {Id: 2, Name: "Test2", AccountId: 1, Positions: []model.Position{{Id: 2, Symbol: "YNDX", Quantity: 5, PlacementTime: time.Time{}}}},
		3: {Id: 3, Name: "Test", AccountId: 2, Positions: []model.Position{{Id: 3, Symbol: "SBER", Quantity: 50, PlacementTime: time.Time{}}}},
	}
}

func assetList() map[int32]*model.Asset {
	return map[int32]*model.Asset{
		1: {Code: "OZON"},
		2: {Code: "YNDX"},
		3: {Code: "SBER"},
	}
}

func portfolioValueList() []*model.PortfolioValue {
	return []*model.PortfolioValue{
		{PortfolioId: 1, Value: 300.0, CalculationTime: time.Now()},
		{PortfolioId: 2, Value: 500.0, CalculationTime: time.Now()},
		{PortfolioId: 3, Value: 900.0, CalculationTime: time.Now()},
	}
}

func TestGetForAccount(t *testing.T) {
	portfolioRepo := store.MockPortfolioRepo{Values: portfolioList()}
	assetRepo := store.MockAssetRepo{Values: assetList()}
	portfolioValueRepo := store.MockPortfolioValueRepo{Values: portfolioValueList()}
	marketClient := mockMarketClient{}

	srv := NewPortfolioService(&portfolioRepo, &assetRepo, &portfolioValueRepo, marketClient)

	portfolioList, err := srv.GetAccountPortfolio(1)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	expected := []model.Portfolio{
		{Id: 1, Name: "Test", AccountId: 1, Positions: []model.Position{{Id: 1, Symbol: "OZON", Quantity: 10, PlacementTime: time.Time{}}}},
		{Id: 2, Name: "Test2", AccountId: 1, Positions: []model.Position{{Id: 2, Symbol: "YNDX", Quantity: 5, PlacementTime: time.Time{}}}},
	}

	if len(portfolioList) != len(expected) {
		t.Errorf("Actual portfolio list len %d not equal to expected len %d", len(portfolioList), len(expected))
	}
}

func TestUpdatePortfolioValues(t *testing.T) {
	portfolioList := portfolioList()
	initValues := portfolioValueList()
	portfolioRepo := store.MockPortfolioRepo{Values: portfolioList}
	assetRepo := store.MockAssetRepo{Values: assetList()}
	portfolioValueRepo := store.MockPortfolioValueRepo{Values: initValues}
	marketClient := mockMarketClient{Values: map[string]*model.StockInfo{
		"OZON": {Code: "OZON", LastPrice: 10.0},
		"YNDX": {Code: "YNDX", LastPrice: 15.0},
		"SBER": {Code: "SBER", LastPrice: 50.0},
	}}

	srv := NewPortfolioService(&portfolioRepo, &assetRepo, &portfolioValueRepo, marketClient)

	err := srv.UpdatePortfolioValues()
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	expectedLen := len(initValues) + len(portfolioList)
	actualLen := len(portfolioValueRepo.Values)
	if actualLen != expectedLen {
		t.Errorf("Actual len of portfolio values %d not equal to expected %d", actualLen, expectedLen)
	}
	expectedValues := map[int32]float64{
		1: 100.0, 2: 75.0, 3: 2500.0,
	}
	for _, p := range portfolioList {
		history, err := portfolioValueRepo.GetPortfolioHistory(p.Id)
		if err != nil {
			t.Errorf("Unexpected error %s", err.Error())
		}
		if math.Abs(history.CurrentValue()-expectedValues[p.Id]) > 0.001 {
			t.Errorf("Actual value %.2f of portfolio %d not equal to expected %.2f", history.CurrentValue(), p.Id, expectedValues[p.Id])
		}
	}
}

func TestCalculateActualValues(t *testing.T) {
	allPortfolioList := make([]*model.Portfolio, 0)
	for _, p := range portfolioList() {
		allPortfolioList = append(allPortfolioList, p)
	}
	marketInfo := map[string]model.StockInfo{
		"OZON": {Code: "OZON", LastPrice: 10.0},
		"YNDX": {Code: "YNDX", LastPrice: 15.0},
		"SBER": {Code: "SBER", LastPrice: 50.0},
	}

	actualValues := calculateActualValues(allPortfolioList, marketInfo)
	expectedValues := map[int32]float64{
		1: 100.0, 2: 75.0, 3: 2500.0,
	}

	for _, av := range actualValues {
		if av.Value != expectedValues[av.PortfolioId] {
			t.Errorf("Actual value %.2f of portfolio %d not equal to expected %.2f", av.Value, av.PortfolioId, expectedValues[av.PortfolioId])
		}
	}
}

func TestNewPortfolio(t *testing.T) {
	portfolioRepo := store.MockPortfolioRepo{Values: portfolioList()}
	assetRepo := store.MockAssetRepo{Values: assetList()}
	portfolioValueRepo := store.MockPortfolioValueRepo{Values: portfolioValueList()}
	marketClient := mockMarketClient{}

	srv := NewPortfolioService(&portfolioRepo, &assetRepo, &portfolioValueRepo, marketClient)

	accountId := int32(1)
	name := "New"
	req := model.NewPortfolioRequest{AccountId: accountId, Name: name}
	p, err := srv.NewPortfolio(req)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	if p == nil {
		t.Errorf("Expected not nil portfolio")
	}
	if p.Id <= 0 {
		t.Errorf("Expected portfolio id > 0, got %d", p.Id)
	}
	if p.AccountId != accountId {
		t.Errorf("Expected account id %d, got %d", accountId, p.AccountId)
	}
	if p.Name != name {
		t.Errorf("Expected portfolio name %s, got %s", name, p.Name)
	}
	if p.Positions != nil && len(p.Positions) > 0 {
		t.Errorf("Expected empty position list for new portfolio")
	}
}

func TestNewPortfolioError(t *testing.T) {
	portfolioRepo := store.MockPortfolioRepo{Values: portfolioList()}
	assetRepo := store.MockAssetRepo{Values: assetList()}
	portfolioValueRepo := store.MockPortfolioValueRepo{Values: portfolioValueList()}
	marketClient := mockMarketClient{}

	srv := NewPortfolioService(&portfolioRepo, &assetRepo, &portfolioValueRepo, marketClient)

	accountId := int32(1)
	name := "Test"
	req := model.NewPortfolioRequest{AccountId: accountId, Name: name}
	_, err := srv.NewPortfolio(req)
	if err == nil || !err.IsType(ex.ConflictError) {
		t.Errorf("Expect conflict error, got %v", err)
	}
}

func TestNewPosition(t *testing.T) {
	portfolioRepo := store.MockPortfolioRepo{Values: portfolioList()}
	assetRepo := store.MockAssetRepo{Values: assetList()}
	portfolioValueRepo := store.MockPortfolioValueRepo{Values: portfolioValueList()}
	marketClient := mockMarketClient{}

	srv := NewPortfolioService(&portfolioRepo, &assetRepo, &portfolioValueRepo, marketClient)

	newSymbol := "OZON"
	newQnt := int32(15)
	req := model.NewPositionRequest{PortfolioId: 3, Symbol: newSymbol, Quantity: newQnt}
	pos, err := srv.PlacePosition(req)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	updatedPortfolio, err := srv.GetById(3)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}
	newPositionList := updatedPortfolio.Positions
	if len(newPositionList) != 2 {
		t.Errorf("Expected %d elements of position list, got %d", 2, len(newPositionList))
	}
	existsNew := false
	for _, p := range updatedPortfolio.Positions {
		if p.Id == pos.Id {
			existsNew = true
		}
	}
	if !existsNew {
		t.Errorf("Expected new position in portfolio")
	}
}
