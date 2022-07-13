package store

import (
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/ex"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
	"math/rand"
)

type MockAccountRepo struct {
	Values map[int32]*model.Account
}

func (m MockAccountRepo) GetAccountById(id int32) (*model.Account, *ex.AppError) {
	if account, ok := m.Values[id]; ok {
		return account, nil
	} else {
		return nil, ex.NewNotFoundError("Not found account")
	}
}

func (m *MockAccountRepo) CreateAccount(account model.Account) (*model.Account, *ex.AppError) {
	id := rand.Int31()
	account.Id = id
	m.Values[id] = &account
	return m.Values[id], nil
}

func (m MockAccountRepo) FindAccount(messenger model.Messenger, messengerId string) (*model.Account, *ex.AppError) {
	var acc *model.Account
	for _, a := range m.Values {
		if a.Messenger == messenger && a.MessengerId == messengerId {
			acc = a
		}
	}
	if acc == nil {
		return nil, ex.NewNotFoundError("Not found account")
	}
	return acc, nil
}

type MockPortfolioRepo struct {
	Values map[int32]*model.Portfolio
}

func (m MockPortfolioRepo) FindByAccount(accountId int32) ([]*model.Portfolio, *ex.AppError) {
	result := make([]*model.Portfolio, 0)
	for k := range m.Values {
		if m.Values[k].AccountId == accountId {
			result = append(result, m.Values[k])
		}
	}
	return result, nil
}

func (m *MockPortfolioRepo) SavePortfolio(portfolio model.Portfolio) (*model.Portfolio, *ex.AppError) {
	id := rand.Int31()
	portfolio.Id = id
	m.Values[id] = &portfolio
	return &portfolio, nil
}

func (m MockPortfolioRepo) GetById(i int32) (*model.Portfolio, *ex.AppError) {
	return m.Values[i], nil
}

func (m MockPortfolioRepo) GetAll() ([]*model.Portfolio, *ex.AppError) {
	result := make([]*model.Portfolio, 0)
	for k := range m.Values {
		result = append(result, m.Values[k])
	}
	return result, nil
}

func (m MockPortfolioRepo) ExistName(accountId int32, name string) (bool, *ex.AppError) {
	for k := range m.Values {
		v := m.Values[k]
		if v.AccountId == accountId && v.Name == name {
			return true, nil
		}
	}
	return false, nil
}

type MockPortfolioValueRepo struct {
	Values []*model.PortfolioValue
}

func (m MockPortfolioValueRepo) GetPortfolioHistory(portfolioId int32) (*model.PortfolioValueHistory, *ex.AppError) {
	result := make([]model.PortfolioValue, 0)
	for _, v := range m.Values {
		if v.PortfolioId == portfolioId {
			result = append(result, *v)
		}
	}
	return model.NewPortfolioValueHistory(result), nil
}

func (m *MockPortfolioValueRepo) SavePortfolioValue(value model.PortfolioValue) *ex.AppError {
	m.Values = append(m.Values, &value)
	return nil
}

type MockAssetRepo struct {
	Values map[int32]*model.Asset
}

func (m *MockAssetRepo) SaveAsset(asset model.Asset) *ex.AppError {
	id := rand.Int31()
	m.Values[id] = &asset
	return nil
}
