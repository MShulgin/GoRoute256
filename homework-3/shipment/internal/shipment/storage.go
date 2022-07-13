package shipment

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx/types"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/db"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/ex"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/pkg/shipment"
	"time"
)

type Storage interface {
	UpdateShipment(shipmentId string, ship shipment.Shipment) *ex.AppError
	GetShipment(shipmentId string) (*shipment.Shipment, *ex.AppError)
	GetShipmentsByOrder(orderId shipment.OrderId) ([]shipment.Shipment, *ex.AppError)
	SaveShipment(ship shipment.Shipment) (*shipment.Shipment, *ex.AppError)
}

type PgStorage struct {
	Cluster *db.PgCluster
}

type shipmentRecord struct {
	Id            string         `db:"id"`
	OrderId       string         `db:"order_id"`
	SellerId      int64          `db:"seller_id"`
	Units         types.JSONText `db:"units"`
	DestinationId int64          `db:"destination_id"`
	Status        int32          `db:"status"`
	CreatedTime   time.Time      `db:"created_time"`
}

func (r *shipmentRecord) ToShipment() (*shipment.Shipment, error) {
	var units []shipment.Unit
	err := r.Units.Unmarshal(&units)
	if err != nil {
		return nil, err
	}

	s := shipment.Shipment{
		Id:            r.Id,
		OrderId:       shipment.OrderId(r.OrderId),
		SellerId:      r.SellerId,
		Units:         units,
		DestinationId: r.DestinationId,
		Status:        shipment.Status(r.Status),
		CreatedTime:   r.CreatedTime,
	}
	return &s, nil
}

func (store *PgStorage) GetShipmentsByOrder(orderId shipment.OrderId) ([]shipment.Shipment, *ex.AppError) {
	shard, err := store.Cluster.GetShard(orderId.CustomerId())
	if err != nil {
		logger.Error("failed to get shard: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	records := make([]shipmentRecord, 0)
	query := `SELECT id, order_id, seller_id, units::text, destination_id, status, created_time 
			  FROM shipment WHERE order_id = $1`
	if err = shard.Select(&records, query, string(orderId)); err != nil {
		logger.Error("failed to get shard: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	ships := make([]shipment.Shipment, 0, len(records))
	for _, r := range records {
		if s, err := r.ToShipment(); err == nil {
			ships = append(ships, *s)
		} else {
			logger.Error("failed to scan shipment from record: " + err.Error())
			return nil, ex.NewUnexpectedError("Unexpected database error")
		}
	}
	return ships, nil
}

func (store *PgStorage) SaveShipment(s shipment.Shipment) (*shipment.Shipment, *ex.AppError) {
	shard, err := store.Cluster.GetShard(s.OrderId.CustomerId())
	unitsJson, err := json.Marshal(&s.Units)
	if err != nil {
		logger.Error("failed to convert shipment units: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	rows := shard.QueryRow("INSERT INTO shipment VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6) RETURNING id",
		s.OrderId, s.SellerId, string(unitsJson), s.DestinationId, int32(s.Status), s.CreatedTime)
	if rows == nil {
		logger.Error("now rows was inserted")
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	var id string
	if err := rows.Scan(&id); err != nil {
		logger.Error("failed to scan shipment id: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	s.Id = id
	return &s, nil
}

func (store *PgStorage) GetShipment(shipmentId string) (*shipment.Shipment, *ex.AppError) {
	var record shipmentRecord
	query := `SELECT id, order_id, seller_id, units, destination_id, status, created_time FROM shipment WHERE id = $1`
	if err := store.Cluster.QueryRow(&record, query, shipmentId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ex.NewNotFoundError(fmt.Sprintf("Not found shipment: shipmentId='%s'", shipmentId))
		}
		logger.Error("failed to scan shipment id: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	ship, err := record.ToShipment()
	if err != nil {
		logger.Error("failed to scan shipment from record: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	return ship, nil
}

func (store *PgStorage) UpdateShipment(shipmentId string, shipment shipment.Shipment) *ex.AppError {
	shard, err := store.Cluster.GetShard(shipment.OrderId.CustomerId())
	if err != nil {
		logger.Error("failed to get shard: " + err.Error())
		return ex.NewUnexpectedError("Unexpected database error")
	}
	exec, err := shard.Exec("UPDATE shipment SET status = $1 WHERE id = $2",
		int32(shipment.Status), shipmentId)
	if err != nil {
		logger.Error("failed to update shipment: " + err.Error())
		return ex.NewUnexpectedError("Unexpected database error")
	}
	if r, err := exec.RowsAffected(); err == nil {
		if r == 0 {
			return ex.NewNotFoundError(fmt.Sprintf("Not found shipment: shipmentId='%s'", shipmentId))
		}
	} else {
		logger.Error("failed to get affected rows: " + err.Error())
		return ex.NewUnexpectedError("Unexpected database error")
	}

	return nil
}
