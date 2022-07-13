package post

import (
	"github.com/jmoiron/sqlx"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/ex"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
)

type Storage interface {
	SaveDelivery(d IncomeDelivery) *ex.AppError
	RemoveDelivery(shipmentId string) *ex.AppError
}

type PgStorage struct {
	Db *sqlx.DB
}

func (s PgStorage) SaveDelivery(d IncomeDelivery) *ex.AppError {
	_, err := s.Db.Exec("INSERT INTO income_delivery (post_id, shipment_id, created_time) VALUES ($1, $2, $3)",
		d.PostId, d.ShipmentId, d.CreatedTime)
	if err != nil {
		logger.Error("failed to save income delivery: " + err.Error())
		return ex.NewUnexpectedError("Unexpected database error")
	}
	return nil
}

func (s PgStorage) RemoveDelivery(shipmentId string) *ex.AppError {
	_, err := s.Db.Exec("DELETE FROM income_delivery WHERE shipment_id = $1", shipmentId)
	if err != nil {
		logger.Error("failed to delete income delivery: " + err.Error())
		return ex.NewUnexpectedError("Unexpected database error")
	}
	return nil
}
