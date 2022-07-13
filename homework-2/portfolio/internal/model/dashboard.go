package model

import "math"

type PortfolioValueInfo struct {
	PortfolioId   int32
	PortfolioName string
	Value         float64
}

func (curr PortfolioValueInfo) Cmp(that PortfolioValueInfo) bool {
	if curr.PortfolioId != that.PortfolioId {
		return false
	}
	if curr.PortfolioName != that.PortfolioName {
		return false
	}
	if math.Abs(curr.Value-that.Value) > 0.001 {
		return false
	}
	return true
}

type AccountDashboard struct {
	AccountId          int32
	TotalValue         float64
	PortfolioValueList []PortfolioValueInfo
}
