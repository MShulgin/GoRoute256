package offer

import (
	"fmt"
	"github.com/Shopify/sarama"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/event"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/ex"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/kafka"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/pkg/shipment"
)

const (
	cancelDelivery = "cancel_delivery"
	commitDelivery = "commit_delivery"
)

type Service struct {
	store         Storage
	kafkaProducer sarama.SyncProducer
}

func NewService(store Storage, kafkaProducer sarama.SyncProducer) Service {
	return Service{store: store, kafkaProducer: kafkaProducer}
}

func (s *Service) generateOfferId(sellerId, productId int64) (string, *ex.AppError) {
	id, err := s.store.NextId()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d-%d-%d", id, sellerId, productId), nil
}

func (s *Service) NewOffer(productId int64, sellerId int64, initStock int64, price float64) (*Offer, *ex.AppError) {
	offerId, err := s.generateOfferId(sellerId, productId)
	if err != nil {
		return nil, err
	}
	offer := Offer{
		Id:        offerId,
		SellerId:  sellerId,
		ProductId: productId,
		Stock:     initStock,
		Reserved:  0,
		Price:     price,
	}
	if err = s.store.SaveOffer(offer); err != nil {
		return nil, err
	}
	return &offer, nil
}

func (s *Service) UpdatePrice(offerId string, newPrice float64) (*Offer, *ex.AppError) {
	updatePriceFn := func(o Offer) Offer {
		o.Price = newPrice
		return o
	}
	return s.store.UpdateOffer(offerId, updatePriceFn)
}

func (s *Service) UpdateStock(offerId string, newStock int64) (*Offer, *ex.AppError) {
	updateStockFn := func(o Offer) Offer {
		o.Stock = newStock
		return o
	}
	return s.store.UpdateOffer(offerId, updateStockFn)
}

func (s *Service) GetPrice(offerId string) (float64, *ex.AppError) {
	off, err := s.store.GetOffer(offerId)
	if err != nil {
		return 0, err
	}

	return off.Price, nil
}

func (s *Service) UpdateReserve(offerId string, newReserve int64) (*Offer, *ex.AppError) {
	updateReserveFn := func(o Offer) Offer {
		o.Reserved = newReserve
		return o
	}
	return s.store.UpdateOffer(offerId, updateReserveFn)
}

func (s *Service) RemoveReserve(shipment shipment.Shipment) *ex.AppError {
	offers := make(map[string]int64)
	for _, u := range shipment.Units {
		offers[u.OfferId] = u.Count
	}
	logger.Info(fmt.Sprintf("Remove reserve for shipment: shipmentId='%s'", shipment.Id))
	if err := s.store.RemoveReserved(offers); err != nil {
		s.sendDeliveryCancel(shipment.Id)
		return err
	}
	//TODO: Send commits from success db table
	if err := s.sendDeliveryCommit(shipment); err != nil {
		return err
	}

	return nil
}

func (s *Service) sendDeliveryCommit(shipment shipment.Shipment) *ex.AppError {
	if err := kafka.SendJsonMessage(s.kafkaProducer, commitDelivery, &shipment); err != nil {
		logger.Error("failed to send msg to kafka" + err.Error())
		return ex.NewUnexpectedError("Unexpected kafka error")
	}
	return nil
}

func (s *Service) sendDeliveryCancel(shipmentId string) {
	cancelMsg := &event.DeliveryCancel{ShipmentId: shipmentId}
	if err := kafka.SendJsonMessage(s.kafkaProducer, cancelDelivery, cancelMsg); err != nil {
		logger.Error("failed to send msg to kafka" + err.Error())
	}
}
