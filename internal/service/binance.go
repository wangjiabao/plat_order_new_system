// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"plat_order/internal/model/entity"
)

type (
	IBinance interface {
		// GetBinancePositionSide 获取账户信息
		GetBinancePositionSide(apiK, apiS string) string
		// GetLatestPrice 获取价格
		GetLatestPrice(symbol string) string
		// GetWalletInfo 获取钱包信息
		GetWalletInfo(apiK, apiS string) []*entity.WalletInfo
		// GetBinanceInfo 获取账户信息
		GetBinanceInfo(apiK, apiS string) string
		RequestBinancePositionSide(positionSide string, apiKey string, secretKey string) (error, string, bool)
		// GetBinanceFuturesPairs 获取 Binance U 本位合约交易对信息
		GetBinanceFuturesPairs() ([]*entity.BinanceSymbolInfo, error)
		// RequestBinanceOrder 请求下单
		RequestBinanceOrder(symbol string, side string, orderType string, positionSide string, quantity string, apiKey string, secretKey string, reduceOnly bool) (*entity.BinanceOrder, *entity.BinanceOrderInfo, error)
		// GetBinancePositionInfo 获取账户信息
		GetBinancePositionInfo(apiK, apiS string) []*entity.BinancePosition
		// CreateListenKey creates a new ListenKey for user data stream
		CreateListenKey(apiKey string) error
		// RenewListenKey renews the ListenKey for user data stream
		RenewListenKey(apiKey string) error
		// ConnectWebSocket safely connects to the WebSocket and updates conn
		ConnectWebSocket() error
	}
)

var (
	localBinance IBinance
)

func Binance() IBinance {
	if localBinance == nil {
		panic("implement not found for interface IBinance, forgot register?")
	}
	return localBinance
}

func RegisterBinance(i IBinance) {
	localBinance = i
}
