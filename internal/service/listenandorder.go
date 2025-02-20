// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"plat_order/internal/model/entity"
)

type (
	IListenAndOrder interface {
		// SetSymbol 更新symbol
		SetSymbol(ctx context.Context) (err error)
		// PullAndSetBaseMoneyNewGuiTuAndUser 拉取binance保证金数据
		PullAndSetBaseMoneyNewGuiTuAndUser(ctx context.Context)
		// PullAndSetTraderUserPositionSide 获取并更新持仓方向
		PullAndSetTraderUserPositionSide(ctx context.Context) (err error)
		// SetUser 初始化用户
		SetUser(ctx context.Context) (err error)
		// HandleBothPositions 处理平仓
		HandleBothPositions(ctx context.Context)
		// OrderAtPlat 在平台下单
		OrderAtPlat(ctx context.Context, doValue *entity.DoValue)
		// Run 监控仓位 pulls binance data and orders
		Run(ctx context.Context)
		// SetPositionSide set position side
		SetPositionSide(apiKey, apiSecret string) (uint64, string)
		// GetSystemUserNum get user num
		GetSystemUserNum(ctx context.Context) map[string]float64
		// CreateUser set user num
		CreateUser(ctx context.Context, address, apiKey, apiSecret, plat string, needInit uint64, num float64) error
		// SetSystemUserNum set user num
		SetSystemUserNum(ctx context.Context, apiKey string, num float64) error
		// SetApiStatus set user api status
		SetApiStatus(ctx context.Context, apiKey string, status uint64, init uint64) uint64
		// SetUseNewSystem set user num
		SetUseNewSystem(ctx context.Context, apiKey string, useNewSystem uint64) error
		// GetSystemUserPositions get user positions
		GetSystemUserPositions(ctx context.Context, apiKey string) map[string]float64
		// GetBinanceUserPositions get binance user positions
		GetBinanceUserPositions(ctx context.Context, apiKey string) map[string]string
		// CloseBinanceUserPositions close binance user positions
		CloseBinanceUserPositions(ctx context.Context) uint64
		// SetSystemUserPosition set user positions
		SetSystemUserPosition(ctx context.Context, system uint64, allCloseGate uint64, apiKey string, symbol string, side string, positionSide string, num float64) uint64
	}
)

var (
	localListenAndOrder IListenAndOrder
)

func ListenAndOrder() IListenAndOrder {
	if localListenAndOrder == nil {
		panic("implement not found for interface IListenAndOrder, forgot register?")
	}
	return localListenAndOrder
}

func RegisterListenAndOrder(i IListenAndOrder) {
	localListenAndOrder = i
}
