package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/kafka"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/pkg/shipment"
	"io"
	"net/http"
)

var (
	brokers     = []string{"localhost:9095"}
	shipmentApi = "http://localhost:8085"
)

const (
	newShipmentDelivery = "new_shipment_delivery"
)

var kafkaProducer sarama.SyncProducer

func main() {
	kp, err := kafka.NewSyncProducer(brokers)
	if err != nil {
		panic(err)
	}
	kafkaProducer = kp

	s1Req := shipment.NewShipmentRequest{
		OrderId:  "1-1",
		SellerId: 1,
		DestId:   1,
		Units:    []shipment.Unit{{OfferId: "1-1-1", Count: 1}},
	}
	s1, err := submitNewShipment(s1Req)
	if err != nil {
		panic(err)
	}
	if err = sendNewDelivery(s1); err != nil {
		panic(err)
	}

	s2Req := shipment.NewShipmentRequest{
		OrderId:  "2-1",
		SellerId: 1,
		DestId:   1,
		Units:    []shipment.Unit{{OfferId: "2-1-2", Count: 5}},
	}
	s2, err := submitNewShipment(s2Req)
	if err != nil {
		panic(err)
	}
	if err = sendNewDelivery(s2); err != nil {
		panic(err)
	}

}

func submitNewShipment(req shipment.NewShipmentRequest) (*shipment.Shipment, error) {
	reqJson, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	r2, err := http.Post(fmt.Sprintf("%s/api/shipment", shipmentApi),
		"application/json", bytes.NewBuffer(reqJson))
	if err != nil {
		return nil, err
	}
	if r2.StatusCode != http.StatusCreated {
		return nil, errors.New(fmt.Sprintf("not expected status = %d", r2.StatusCode))
	}

	body, err := io.ReadAll(r2.Body)
	if err != nil {
		return nil, err
	}
	var s2 shipment.Shipment
	if err = json.Unmarshal(body, &s2); err != nil {
		return nil, err
	}
	return &s2, nil
}

func sendNewDelivery(s *shipment.Shipment) error {
	msg, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err = kafka.SendMessage(kafkaProducer, newShipmentDelivery, msg); err != nil {
		return err
	}
	return nil
}
