package dto

type OfferPrice struct {
	OfferId string  `json:"offerId"`
	Price   float64 `json:"price"`
}

type NewOfferRequest struct {
	SellerId  int64   `json:"sellerId"`
	ProductId int64   `json:"productId"`
	Stock     int64   `json:"stock"`
	Price     float64 `json:"price"`
}
