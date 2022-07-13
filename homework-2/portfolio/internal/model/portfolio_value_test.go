package model

import (
	"testing"
	"time"
)

var startDate = time.Unix(1640995200, 0)

var history = []PortfolioValue{
	{PortfolioId: 1, Value: 300.0, CalculationTime: startDate},
	{PortfolioId: 1, Value: 400.0, CalculationTime: startDate.Add(1 * 24 * time.Hour)},
	{PortfolioId: 1, Value: 500.0, CalculationTime: startDate.Add(2 * 24 * time.Hour)},
	{PortfolioId: 1, Value: 600.0, CalculationTime: startDate.Add(3 * 24 * time.Hour)},
}

func TestNewPortfolioValueHistory(t *testing.T) {
	unsorted := []PortfolioValue{history[2], history[1], history[0], history[3]}
	valueHistory := NewPortfolioValueHistory(unsorted)

	var current []PortfolioValue = *valueHistory
	expected := history
	for i, _ := range current {
		if current[i] != expected[i] {
			t.Errorf("Position %v not equal to expected %v", current[i], expected[i])
		}
	}
}

func TestPortfolioValueHistory_CurrentValue(t *testing.T) {
	valueHistory := NewPortfolioValueHistory(history)
	current := valueHistory.CurrentValue()
	expected := history[len(history)-1].Value
	if expected != current {
		t.Errorf("Current value %.2f not equal to expected %.2f", current, expected)
	}
}
