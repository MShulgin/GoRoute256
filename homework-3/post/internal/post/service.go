package post

import (
	"fmt"
	"github.com/Shopify/sarama"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/event"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/ex"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/kafka"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
	"gitlab.ozon.dev/MShulgin/homework-3/post/internal/topic"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/pkg/shipment"
	"time"
)

type Service struct {
	store         Storage
	kafkaProducer sarama.SyncProducer
}

func NewService(store Storage, kafkaProducer sarama.SyncProducer) Service {
	return Service{store: store, kafkaProducer: kafkaProducer}
}

func (s *Service) RegisterDelivery(shipment shipment.Shipment) *ex.AppError {
	d := IncomeDelivery{
		PostId:      shipment.DestinationId,
		ShipmentId:  shipment.Id,
		CreatedTime: time.Now(),
	}
	logger.Info(fmt.Sprintf("Save delivery: shipmentId=%s, postId=%d", d.ShipmentId, d.PostId))
	if err := s.store.SaveDelivery(d); err != nil {
		s.sendDeliveryCancel(shipment.Id)
		return err
	}
	if err := kafka.SendJsonMessage(s.kafkaProducer, topic.RemoveReserve, &shipment); err != nil {
		logger.Error("failed to send remove reserved message: " + err.Error())
		s.sendDeliveryCancel(shipment.Id)
	}
	return nil
}

func (s *Service) sendDeliveryCancel(shipmentId string) {
	cancelMsg := event.DeliveryCancel{ShipmentId: shipmentId}
	if err := kafka.SendJsonMessage(s.kafkaProducer, topic.CancelDelivery, &cancelMsg); err != nil {
		logger.Error("failed to send msg to kafka: " + err.Error())
	}
}

func (s *Service) DeleteDelivery(shipmentId string) *ex.AppError {
	logger.Info(fmt.Sprintf("Remove delivery: shipmentId=%s", shipmentId))
	return s.store.RemoveDelivery(shipmentId)
}
