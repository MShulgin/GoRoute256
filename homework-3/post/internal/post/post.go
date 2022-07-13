package post

import "time"

type IncomeDelivery struct {
	PostId      int64
	ShipmentId  string
	CreatedTime time.Time
}
