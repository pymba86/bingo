package engine

import (
	"github.com/pymba86/bingo/pkg/fixedpoint"
	"github.com/pymba86/bingo/pkg/types"
)

type ExchangeSession struct {
	Name         string             `json:"name,omitempty" yaml:"name,omitempty"`
	ExchangeName types.ExchangeName `json:"exchange" yaml:"exchange"`
	EnvVarPrefix string             `json:"envVarPrefix" yaml:"envVarPrefix"`
	Key          string             `json:"key,omitempty" yaml:"key,omitempty"`
	Secret       string             `json:"secret,omitempty" yaml:"secret,omitempty"`
	SubAccount   string             `json:"subAccount,omitempty" yaml:"subAccount,omitempty"`

	Withdrawal   bool             `json:"withdrawal,omitempty" yaml:"withdrawal,omitempty"`
	MakerFeeRate fixedpoint.Value `json:"makerFeeRate,omitempty" yaml:"makerFeeRate,omitempty"`
	TakerFeeRate fixedpoint.Value `json:"takerFeeRate,omitempty" yaml:"takerFeeRate,omitempty"`

	PublicOnly           bool   `json:"publicOnly,omitempty" yaml:"publicOnly"`
	Margin               bool   `json:"margin,omitempty" yaml:"margin"`
	IsolatedMargin       bool   `json:"isolatedMargin,omitempty" yaml:"isolatedMargin,omitempty"`
	IsolatedMarginSymbol string `json:"isolatedMarginSymbol,omitempty" yaml:"isolatedMarginSymbol,omitempty"`

	Account *types.Account `json:"-" yaml:"-"`

	IsInitialized bool `json:"-" yaml:"-"`

	Exchange types.Exchange `json:"-" yaml:"-"`

	UserDataStream   types.Stream `json:"-" yaml:"-"`
	MarketDataStream types.Stream `json:"-" yaml:"-"`

	OrderExecutor *ExchangeOrderExecutor `json:"orderExecutor,omitempty" yaml:"orderExecutor,omitempty"`

	Subscriptions map[types.Subscription]types.Subscription `json:"-" yaml:"-"`

}
