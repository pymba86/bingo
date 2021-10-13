package engine

import (
	"context"
	"github.com/pymba86/bingo/pkg/types"
)

type OrderExecutor interface {
	SubmitOrders(ctx context.Context, orders ...types.SubmitOrder) (createdOrders types.OrderSlice, err error)
	OnTradeUpdate(cb func(trade types.Trade))
	OnOrderUpdate(cb func(order types.Order))
	EmitTradeUpdate(trade types.Trade)
	EmitOrderUpdate(order types.Order)
}

type OrderExecutionRouter interface {
	// SubmitOrdersTo submit order to a specific exchange Session
	SubmitOrdersTo(ctx context.Context, session string, orders ...types.SubmitOrder) (
		createdOrders types.OrderSlice, err error)
}

type ExchangeOrderExecutionRouter struct {
	Notifiability

	sessions map[string]*ExchangeSession
	executors map[string]OrderExecutor
}

type ExchangeOrderExecutor struct {

	Notifiability `json:"-" yaml:"-"`

	Session *ExchangeSession `json:"-" yaml:"-"`

	tradeUpdateCallbacks []func(trade types.Trade)

	orderUpdateCallbacks []func(order types.Order)
}

func (e *ExchangeOrderExecutor) SubmitOrders(ctx context.Context, orders ...types.SubmitOrder) (
	createdOrders types.OrderSlice, err error) {
	panic("implement me")
}

func (e *ExchangeOrderExecutor) OnTradeUpdate(cb func(trade types.Trade)) {
	e.tradeUpdateCallbacks = append(e.tradeUpdateCallbacks, cb)
}

func (e *ExchangeOrderExecutor) OnOrderUpdate(cb func(order types.Order)) {
	e.orderUpdateCallbacks = append(e.orderUpdateCallbacks, cb)
}

func (e *ExchangeOrderExecutor) EmitTradeUpdate(trade types.Trade) {
	for _, cb := range e.tradeUpdateCallbacks {
		cb(trade)
	}
}

func (e *ExchangeOrderExecutor) EmitOrderUpdate(order types.Order) {
	for _, cb := range e.orderUpdateCallbacks {
		cb(order)
	}
}
