package market

type Marketdata struct {
	Data [][]interface{} `json:"data"`
}

type MoexStockInfo struct {
	Marketdata Marketdata `json:"marketdata"`
}
