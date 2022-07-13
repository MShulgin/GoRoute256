package offer

type Offer struct {
	Id        string  `json:"id" db:"id"`
	SellerId  int64   `json:"sellerId" db:"seller_id"`
	ProductId int64   `json:"productId" db:"product_id"`
	Price     float64 `json:"price" db:"price"`
	Stock     int64   `json:"stock" db:"stock"`
	Reserved  int64   `json:"reserved" db:"reserved"`
}
