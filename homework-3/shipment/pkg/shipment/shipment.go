package shipment

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Unit struct {
	OfferId string `json:"offerId"`
	Count   int64  `json:"count"`
}

type Shipment struct {
	Id            string    `json:"id" db:"id"`
	OrderId       OrderId   `json:"orderId" db:"order_id"`
	SellerId      int64     `json:"sellerId" db:"seller_id"`
	Units         []Unit    `json:"units" db:"units"`
	DestinationId int64     `json:"destinationId" db:"destination_id"`
	Status        Status    `json:"status" db:"status"`
	CreatedTime   time.Time `json:"createdTime" db:"created_time"`
}

type OrderId string

var orderIdRE = regexp.MustCompile("[0-9]+-[0-9]+")

func OrderIdFromString(id string) (OrderId, bool) {
	if m := orderIdRE.MatchString(id); m {
		return OrderId(id), true
	} else {
		return "", false
	}
}

func (orderId OrderId) CustomerId() string {
	c, _ := orderId.split()
	return c
}

func (orderId OrderId) split() (customerId string, orderNo string) {
	customerId, orderNo, _ = strings.Cut(string(orderId), "-")
	return
}

type Status int32

func (s Status) MarshalJSON() ([]byte, error) {
	asStr := s.String()
	return json.Marshal(&asStr)
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var statusStr string
	err := json.Unmarshal(data, &statusStr)
	if err != nil {
		return err
	}
	p, err := StatusFromString(statusStr)
	if err != nil {
		return err
	}
	*s = p
	return nil
}

const (
	Created Status = iota
	Packing
	AcceptDelivery
	InDelivery
	Delivered
	Received
	Cancelled
)

func (s Status) String() string {
	switch s {
	case Created:
		return "Created"
	case Packing:
		return "Packing"
	case AcceptDelivery:
		return "AcceptDelivery"
	case InDelivery:
		return "InDelivery"
	case Delivered:
		return "Delivered"
	case Received:
		return "Received"
	case Cancelled:
		return "Cancelled"
	default:
		panic("unknown shipment status")
	}
}

func StatusFromString(s string) (Status, error) {
	switch s {
	case "Created":
		return Created, nil
	case "Packing":
		return Packing, nil
	case "AcceptDelivery":
		return AcceptDelivery, nil
	case "InDelivery":
		return InDelivery, nil
	case "Delivered":
		return Delivered, nil
	case "Received":
		return Received, nil
	case "Cancelled":
		return Cancelled, nil
	default:
		return Created, errors.New(fmt.Sprintf("Unknown shipment status: %s", s))
	}
}

type NewShipmentRequest struct {
	OrderId  string `json:"orderId"`
	SellerId int64  `json:"sellerId"`
	DestId   int64  `json:"destinationId"`
	Units    []Unit `json:"units"`
}
