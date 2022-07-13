package market

import (
	"encoding/json"
	"fmt"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/ex"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/logging"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
	"net/http"
	"time"
)

const MOEX_BASE = "https://iss.moex.com"

type MoexStockClient struct {
	client http.Client
}

func NewMoexStockClient() MoexStockClient {
	return MoexStockClient{client: http.Client{Timeout: 10 * time.Second}}
}

func (m MoexStockClient) GetStockInfo(code string) (*model.StockInfo, *ex.AppError) {
	url := fmt.Sprintf("%s/iss/engines/stock/markets/shares/boards/TQBR/securities/%s.json", MOEX_BASE, code)
	resp, err := m.client.Get(url)
	if err != nil {
		logging.Error("Error getting security info: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error in moex client")
	}
	defer resp.Body.Close()

	var moexDto MoexStockInfo
	err = json.NewDecoder(resp.Body).Decode(&moexDto)
	if err != nil {
		logging.Error("Fail to decode json data: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error in moex client")
	}

	stockData := moexDto.Marketdata.Data

	if len(stockData) == 0 {
		return nil, ex.NewNotFoundError(fmt.Sprintf("Not found data for %s in moex", code))
	}
	lastPrice, ok := stockData[0][12].(float64)
	if !ok {
		return nil, ex.NewNotFoundError(fmt.Sprintf("Not found price for %s in moex", code))
	}
	stockInfo := model.StockInfo{Code: code, LastPrice: lastPrice}

	return &stockInfo, nil
}

func (m MoexStockClient) GetMarketInfo() (map[string]model.StockInfo, *ex.AppError) {
	url := fmt.Sprintf("%s/iss/engines/stock/markets/shares/boards/TQBR/securities.json", MOEX_BASE)
	resp, err := m.client.Get(url)
	if err != nil {
		logging.Error("Error getting security list: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error in moex client")
	}
	defer resp.Body.Close()

	var moexDto MoexStockInfo
	err = json.NewDecoder(resp.Body).Decode(&moexDto)
	if err != nil {
		logging.Error("Fail to decode json data: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error in moex client")
	}
	marketData := moexDto.Marketdata.Data
	infoList := make(map[string]model.StockInfo, len(marketData))
	for _, stockData := range marketData {
		code, ok := stockData[0].(string)
		if !ok {
			continue
		}
		lastPrice, ok := stockData[12].(float64)
		if ok {
			infoList[code] = model.StockInfo{Code: code, LastPrice: lastPrice}
		}
	}

	return infoList, nil
}
