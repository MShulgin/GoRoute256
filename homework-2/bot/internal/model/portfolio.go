package model

import "time"

type Account struct {
	Id int32
}

type Portfolio struct {
	Id        int32
	Name      string
	Positions []Position
}

type Position struct {
	Id            int32
	Symbol        string
	Quantity      int32
	PlacementTime time.Time
}

type PortfolioValue struct {
	Id    int32
	Name  string
	Value float64
}

type Dashboard struct {
	TotalValue float64
	ValueList  []PortfolioValue
}
