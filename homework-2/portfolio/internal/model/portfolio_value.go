package model

import (
	"sort"
	"time"
)

type PortfolioValue struct {
	PortfolioId     int32     `db:"portfolio_id"`
	Value           float64   `db:"value"`
	CalculationTime time.Time `db:"calculation_time"`
}

type PortfolioValueHistory []PortfolioValue

func NewPortfolioValueHistory(values []PortfolioValue) *PortfolioValueHistory {
	sort.Slice(values, func(i, j int) bool {
		return values[i].CalculationTime.Before(values[j].CalculationTime)
	})
	return (*PortfolioValueHistory)(&values)
}

func (h PortfolioValueHistory) CurrentValue() float64 {
	last := len(h) - 1
	if last < 0 {
		return 0.0
	}
	return h[last].Value
}
