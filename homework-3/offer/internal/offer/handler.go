package offer

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/ex"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/handler"
	"gitlab.ozon.dev/MShulgin/homework-3/offer/internal/dto"
	"net/http"
)

type Handler struct {
	OfferService Service
}

func (h *Handler) GetOfferPrice(w http.ResponseWriter, r *http.Request) {
	requestVars := mux.Vars(r)
	offerId := requestVars["offerId"]
	if price, err := h.OfferService.GetPrice(offerId); err == nil {
		payload := dto.OfferPrice{OfferId: offerId, Price: price}
		handler.WriteJson(200, w, &payload)
	} else {
		handler.WriteError(w, err)
	}
}

func (h *Handler) SaveOffer(w http.ResponseWriter, r *http.Request) {
	var req dto.NewOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.WriteError(w, ex.NewBadRequestError(err.Error()))
		return
	}
	if offer, err := h.OfferService.NewOffer(req.ProductId, req.SellerId, req.Stock, req.Price); err == nil {
		handler.WriteJson(http.StatusCreated, w, &offer)
	} else {
		handler.WriteError(w, err)
	}
}
