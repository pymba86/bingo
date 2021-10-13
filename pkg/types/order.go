package types

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type MarginOrderSideEffectType string

var (
	SideEffectTypeNoSideEffect MarginOrderSideEffectType = "NO_SIDE_EFFECT"
	SideEffectTypeMarginBuy    MarginOrderSideEffectType = "MARGIN_BUY"
	SideEffectTypeAutoRepay    MarginOrderSideEffectType = "AUTO_REPAY"
)

func (t *MarginOrderSideEffectType) UnmarshalJSON(data []byte) error {
	var s string
	var err = json.Unmarshal(data, &s)
	if err != nil {
		return errors.Wrapf(err, "unable to unmarshal side effect type: %s", data)
	}

	switch strings.ToUpper(s) {

	case string(SideEffectTypeNoSideEffect), "":
		*t = SideEffectTypeNoSideEffect
		return nil

	case string(SideEffectTypeMarginBuy), "BORROW", "MARGINBUY":
		*t = SideEffectTypeMarginBuy
		return nil

	case string(SideEffectTypeAutoRepay), "REPAY", "AUTOREPAY":
		*t = SideEffectTypeAutoRepay
		return nil

	}

	return fmt.Errorf("invalid side effect type: %s", data)
}

type OrderType string

const (
	OrderTypeLimit      OrderType = "LIMIT"
	OrderTypeLimitMaker OrderType = "LIMIT_MAKER"
	OrderTypeMarket     OrderType = "MARKET"
	OrderTypeStopLimit  OrderType = "STOP_LIMIT"
	OrderTypeStopMarket OrderType = "STOP_MARKET"
	OrderTypeIOCLimit   OrderType = "IOC_LIMIT"
)

const NoClientOrderID = "0"

type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusRejected        OrderStatus = "REJECTED"
)

type SubmitOrder struct {
	ClientOrderId string `json:"clientOrderID" db:"client_order_id"`

	Symbol string    `json:"symbol" db:"symbol"`
	Side   SideType  `json:"side" db:"side"`
	Type   OrderType `json:"orderType" db:"order_type"`

	Quantity  float64 `json:"quantity" db:"quantity"`
	Price     float64 `json:"price" db:"price"`
	StopPrice float64 `json:"stopPrice,omitempty" db:"stop_price"`

	Market Market `json:"-" db:"-"`

	StopPriceString string `json:"-"`
	PriceString     string `json:"-"`
	QuantityString  string `json:"-"`

	TimeInForce string `json:"timeInForce,omitempty" db:"time_in_force"` // GTC, IOC, FOK

	GroupID uint32 `json:"groupID,omitempty"`

	MarginSideEffect MarginOrderSideEffectType `json:"marginSideEffect,omitempty"`
}


type Order struct {
	SubmitOrder

	Exchange         ExchangeName `json:"exchange" db:"exchange"`
	GID              uint64       `json:"gid" db:"gid"`
	OrderID          uint64       `json:"orderID" db:"order_id"` // order id
	Status           OrderStatus  `json:"status" db:"status"`
	ExecutedQuantity float64      `json:"executedQuantity" db:"executed_quantity"`
	IsWorking        bool         `json:"isWorking" db:"is_working"`
	CreationTime     Time         `json:"creationTime" db:"created_at"`
	UpdateTime       Time         `json:"updateTime" db:"updated_at"`

	IsMargin   bool `json:"isMargin" db:"is_margin"`
	IsIsolated bool `json:"isIsolated" db:"is_isolated"`
}

func (o Order) Backup() SubmitOrder {
	so := o.SubmitOrder
	so.Quantity = o.Quantity - o.ExecutedQuantity

	// ClientOrderID can not be reused
	so.ClientOrderId = ""
	return so
}