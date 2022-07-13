package model

import (
	"testing"
	"time"
)

var emptyPortfolio = Portfolio{
	Id:        1,
	Name:      "Test",
	AccountId: 1,
	Positions: make([]Position, 0),
}

func TestPortfolio_AddPosition(t *testing.T) {
	portfolio := emptyPortfolio
	initPosition := Position{
		Id:            1,
		Symbol:        "OZON",
		Quantity:      10,
		PlacementTime: time.Now(),
	}
	portfolio.Positions = []Position{initPosition}

	newPosition := Position{
		Id:            0,
		Symbol:        "YNDX",
		Quantity:      50,
		PlacementTime: time.Now(),
	}
	portfolio.AddPosition(newPosition)
	expected := []Position{initPosition, newPosition}
	if len(portfolio.Positions) != len(expected) {
		t.Errorf("Len of actual slice %d not equal to expected %d", len(portfolio.Positions), len(expected))
	}
	for i, _ := range portfolio.Positions {
		if portfolio.Positions[i] != expected[i] {
			t.Errorf("Position %q not equal to expected %q", portfolio.Positions[i], expected[i])
		}
	}
}

func TestPortfolio_DeletePosition(t *testing.T) {
	portfolio := emptyPortfolio
	pos1 := Position{
		Id:            1,
		Symbol:        "OZON",
		Quantity:      10,
		PlacementTime: time.Now(),
	}
	pos2 := Position{
		Id:            2,
		Symbol:        "YNDX",
		Quantity:      10,
		PlacementTime: time.Now(),
	}
	pos3 := Position{
		Id:            3,
		Symbol:        "SBER",
		Quantity:      5,
		PlacementTime: time.Now(),
	}
	portfolio.Positions = []Position{pos1, pos2, pos3}

	portfolio.DeletePosition(2)
	expected := []Position{pos1, pos3}
	if len(portfolio.Positions) != len(expected) {
		t.Errorf("Len of actual slice %d not equal to expected %d", len(portfolio.Positions), len(expected))
	}
	for i, _ := range portfolio.Positions {
		if portfolio.Positions[i] != expected[i] {
			t.Errorf("Position %q not equal to expected %q", portfolio.Positions[i], expected[i])
		}
	}
}
