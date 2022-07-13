package shipment

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/ex"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/handler"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/pkg/shipment"
	"net/http"
)

type Handler struct {
	ShipmentService Service
}

func NewHandler(service Service) Handler {
	return Handler{ShipmentService: service}
}

func (h *Handler) SaveShipment(w http.ResponseWriter, r *http.Request) {
	var req shipment.NewShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := ex.NewBadRequestError(err.Error())
		handler.WriteError(w, appErr)
	}
	if offer, err := h.ShipmentService.NewShipment(req.OrderId, req.SellerId, req.DestId, req.Units); err == nil {
		handler.WriteJson(http.StatusCreated, w, &offer)
	} else {
		handler.WriteError(w, err)
	}
}

func (h *Handler) GetShipment(w http.ResponseWriter, r *http.Request) {
	requestVars := mux.Vars(r)
	shipId := requestVars["shipmentId"]
	if ship, err := h.ShipmentService.GetShipment(shipId); err == nil {
		handler.WriteJson(http.StatusOK, w, &ship)
	} else {
		handler.WriteError(w, err)
	}
}

func (h *Handler) FilterShipments(w http.ResponseWriter, r *http.Request) {
	requestVars := mux.Vars(r)
	orderId := requestVars["orderId"]
	if shipments, err := h.ShipmentService.GetOrderShipments(orderId); err == nil {
		handler.WriteJson(http.StatusOK, w, &shipments)
	} else {
		handler.WriteError(w, err)
	}
}
