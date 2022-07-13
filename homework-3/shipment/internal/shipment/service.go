package shipment

import (
	"fmt"
	"github.com/Shopify/sarama"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/ex"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/kafka"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/internal/topic"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/pkg/shipment"
	"time"
)

type Service struct {
	storage       Storage
	kafkaProducer sarama.SyncProducer
}

func NewService(store Storage, kafkaProducer sarama.SyncProducer) Service {
	return Service{storage: store, kafkaProducer: kafkaProducer}
}

func (srv *Service) NewShipment(orderId string, sellerId int64, destId int64, units []shipment.Unit) (*shipment.Shipment, *ex.AppError) {
	ship := shipment.Shipment{
		OrderId:       shipment.OrderId(orderId),
		SellerId:      sellerId,
		Units:         units,
		DestinationId: destId,
		Status:        shipment.Created,
		CreatedTime:   time.Now(),
	}
	return srv.storage.SaveShipment(ship)
}

func (srv *Service) GetOrderShipments(orderId string) ([]shipment.Shipment, *ex.AppError) {
	id, ok := shipment.OrderIdFromString(orderId)
	if !ok {
		return nil, ex.NewBadRequestError(fmt.Sprintf("Invalid OrderId format: '%s', Accept format: '1-1'", orderId))
	}
	return srv.storage.GetShipmentsByOrder(id)
}

func (srv *Service) GetShipment(shipmentId string) (*shipment.Shipment, *ex.AppError) {
	return srv.storage.GetShipment(shipmentId)
}

func (srv *Service) AcceptDelivery(s shipment.Shipment) *ex.AppError {
	ship, err := srv.storage.GetShipment(s.Id)
	if err != nil {
		return err
	}
	newStatus := shipment.AcceptDelivery
	if ship.Status == shipment.Cancelled {
		return ex.NewBadRequestError(fmt.Sprintf("forbidden to change status of cancelled shipment: shipmentId='%s'", ship.Id))
	}
	if ship.Status >= shipment.AcceptDelivery {
		return ex.NewBadRequestError(fmt.Sprintf("forbidden to transfer shipment status from '%s' to '%s'", ship.Status, newStatus))
	}
	ship.Status = newStatus
	err = srv.storage.UpdateShipment(s.Id, *ship)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Update shipment status: id=%s, status=%s", s.Id, newStatus))

	if err := kafka.SendJsonMessage(srv.kafkaProducer, topic.NewPostDelivery, &ship); err != nil {
		logger.Error("failed to send msg to kafka" + err.Error())
	}

	return nil
}

func (srv *Service) CommitDelivery(s shipment.Shipment) *ex.AppError {
	ship, err := srv.storage.GetShipment(s.Id)
	if err != nil {
		return err
	}
	if ship.Status == shipment.Cancelled {
		return ex.NewBadRequestError(fmt.Sprintf("forbidden to change status of cancelled shipment: shipmentId='%s'", ship.Id))
	}
	newStatus := shipment.InDelivery
	ship.Status = newStatus
	if err := srv.storage.UpdateShipment(s.Id, *ship); err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Update shipment status: id=%s, status=%s", s.Id, newStatus))

	return nil
}

func (srv *Service) CancelDelivery(shipmentId string) *ex.AppError {
	ship, err := srv.storage.GetShipment(shipmentId)
	if err != nil {
		return err
	}
	if ship.Status == shipment.Cancelled {
		return ex.NewBadRequestError(fmt.Sprintf("forbidden to change status of cancelled shipment: shipmentId='%s'", ship.Id))
	}
	newStatus := shipment.Packing
	ship.Status = newStatus
	if err = srv.storage.UpdateShipment(shipmentId, *ship); err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Update shipment status: id=%s, status=%s", shipmentId, newStatus))

	return nil
}
