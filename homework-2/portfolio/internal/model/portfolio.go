package model

import (
	"time"
)

type Portfolio struct {
	Id        int32  `db:"id"`
	Name      string `db:"label"`
	AccountId int32  `db:"account_id"`
	Positions []Position
}

func (p *Portfolio) AddPosition(newPos Position) {
	p.Positions = append(p.Positions, newPos)
}

func (p *Portfolio) DeletePosition(posId int32) {
	idx := 0
	for i, p := range p.Positions {
		if p.Id == posId {
			idx = i
		}
	}
	p.Positions = append(p.Positions[0:idx], p.Positions[idx+1:]...)
}

type Position struct {
	Id            int32     `db:"id"`
	Symbol        string    `db:"code"`
	Quantity      int32     `db:"quantity"`
	PlacementTime time.Time `db:"placement_time"`
}

func NewPosition(symbol string, qnt int32) Position {
	return Position{
		Id:            0,
		Symbol:        symbol,
		Quantity:      qnt,
		PlacementTime: time.Now(),
	}
}

type Asset struct {
	Code string
}

type NewPortfolioRequest struct {
	AccountId int32
	Name      string
}

type NewPositionRequest struct {
	PortfolioId int32
	Symbol      string
	Quantity    int32
}
