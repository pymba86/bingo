package types

import (
	"context"
	"github.com/gorilla/websocket"
)

type Channel string

var BookChannel = Channel("book")

var KLineChannel = Channel("kline")

type Stream interface {
	StandardStreamEventHub

	Subscribe(channel Channel, symbol string, options SubscribeOptions)
	SetPublicOnly()
	Connect(ctx context.Context) error
	Close() error
}

type StandardStream struct {
	ReconnectC chan struct{}

	Subscriptions []Subscription

	startCallbacks []func()

	connectCallbacks []func()

	disconnectCallbacks []func()

	// private trade update callbacks
	tradeUpdateCallbacks []func(trade Trade)

	// private order update callbacks
	orderUpdateCallbacks []func(order Order)

	// balance snapshot callbacks
	balanceSnapshotCallbacks []func(balances BalanceMap)

	balanceUpdateCallbacks []func(balances BalanceMap)

	kLineClosedCallbacks []func(kline KLine)

	kLineCallbacks []func(kline KLine)

	bookUpdateCallbacks []func(book SliceOrderBook)

	bookSnapshotCallbacks []func(book SliceOrderBook)
}

func (stream *StandardStream) Subscribe(channel Channel, symbol string, options SubscribeOptions) {
	stream.Subscriptions = append(stream.Subscriptions, Subscription{
		Channel: channel,
		Symbol:  symbol,
		Options: options,
	})
}

func (stream *StandardStream) Reconnect() {
	select {
	case stream.ReconnectC <- struct{}{}:
	default:
	}
}

func (stream *StandardStream) Dial(url string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	// use the default ping handler
	conn.SetPingHandler(nil)
	return conn, nil
}

func (stream *StandardStream) OnStart(cb func()) {
	stream.startCallbacks = append(stream.startCallbacks, cb)
}

func (stream *StandardStream) EmitStart() {
	for _, cb := range stream.startCallbacks {
		cb()
	}
}

func (stream *StandardStream) OnConnect(cb func()) {
	stream.connectCallbacks = append(stream.connectCallbacks, cb)
}

func (stream *StandardStream) EmitConnect() {
	for _, cb := range stream.connectCallbacks {
		cb()
	}
}

func (stream *StandardStream) OnDisconnect(cb func()) {
	stream.disconnectCallbacks = append(stream.disconnectCallbacks, cb)
}

func (stream *StandardStream) EmitDisconnect() {
	for _, cb := range stream.disconnectCallbacks {
		cb()
	}
}

func (stream *StandardStream) OnTradeUpdate(cb func(trade Trade)) {
	stream.tradeUpdateCallbacks = append(stream.tradeUpdateCallbacks, cb)
}

func (stream *StandardStream) EmitTradeUpdate(trade Trade) {
	for _, cb := range stream.tradeUpdateCallbacks {
		cb(trade)
	}
}

func (stream *StandardStream) OnOrderUpdate(cb func(order Order)) {
	stream.orderUpdateCallbacks = append(stream.orderUpdateCallbacks, cb)
}

func (stream *StandardStream) EmitOrderUpdate(order Order) {
	for _, cb := range stream.orderUpdateCallbacks {
		cb(order)
	}
}

func (stream *StandardStream) OnBalanceSnapshot(cb func(balances BalanceMap)) {
	stream.balanceSnapshotCallbacks = append(stream.balanceSnapshotCallbacks, cb)
}

func (stream *StandardStream) EmitBalanceSnapshot(balances BalanceMap) {
	for _, cb := range stream.balanceSnapshotCallbacks {
		cb(balances)
	}
}

func (stream *StandardStream) OnBalanceUpdate(cb func(balances BalanceMap)) {
	stream.balanceUpdateCallbacks = append(stream.balanceUpdateCallbacks, cb)
}

func (stream *StandardStream) EmitBalanceUpdate(balances BalanceMap) {
	for _, cb := range stream.balanceUpdateCallbacks {
		cb(balances)
	}
}

func (stream *StandardStream) OnKLineClosed(cb func(kline KLine)) {
	stream.kLineClosedCallbacks = append(stream.kLineClosedCallbacks, cb)
}

func (stream *StandardStream) EmitKLineClosed(kline KLine) {
	for _, cb := range stream.kLineClosedCallbacks {
		cb(kline)
	}
}

func (stream *StandardStream) OnKLine(cb func(kline KLine)) {
	stream.kLineCallbacks = append(stream.kLineCallbacks, cb)
}

func (stream *StandardStream) EmitKLine(kline KLine) {
	for _, cb := range stream.kLineCallbacks {
		cb(kline)
	}
}

func (stream *StandardStream) OnBookUpdate(cb func(book SliceOrderBook)) {
	stream.bookUpdateCallbacks = append(stream.bookUpdateCallbacks, cb)
}

func (stream *StandardStream) EmitBookUpdate(book SliceOrderBook) {
	for _, cb := range stream.bookUpdateCallbacks {
		cb(book)
	}
}

func (stream *StandardStream) OnBookSnapshot(cb func(book SliceOrderBook)) {
	stream.bookSnapshotCallbacks = append(stream.bookSnapshotCallbacks, cb)
}

func (stream *StandardStream) EmitBookSnapshot(book SliceOrderBook) {
	for _, cb := range stream.bookSnapshotCallbacks {
		cb(book)
	}
}

type StandardStreamEventHub interface {
	OnStart(cb func())

	OnConnect(cb func())

	OnDisconnect(cb func())

	OnTradeUpdate(cb func(trade Trade))

	OnOrderUpdate(cb func(order Order))

	OnBalanceSnapshot(cb func(balances BalanceMap))

	OnBalanceUpdate(cb func(balances BalanceMap))

	OnKLineClosed(cb func(kline KLine))

	OnKLine(cb func(kline KLine))

	OnBookUpdate(cb func(book SliceOrderBook))

	OnBookSnapshot(cb func(book SliceOrderBook))
}

type SubscribeOptions struct {
	Interval string `json:"interval,omitempty"`
	Depth    string `json:"depth,omitempty"`
}

func (o SubscribeOptions) String() string {
	if len(o.Interval) > 0 {
		return o.Interval
	}

	return o.Depth
}

type Subscription struct {
	Symbol string `json:"symbol"`
	Channel Channel `json:"channel"`
	Options SubscribeOptions `json:"options"`
}
