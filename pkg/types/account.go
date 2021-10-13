package types

import "github.com/pymba86/bingo/pkg/fixedpoint"

type Balance struct {
	Currency  string           `json:"currency"`
	Available fixedpoint.Value `json:"available"`
	Locked    fixedpoint.Value `json:"locked"`
}

func (b Balance) Total() fixedpoint.Value {
	return b.Available + b.Locked
}

type BalanceMap map[string]Balance

type Account struct {
	MakerCommission fixedpoint.Value `json:"makerCommission,omitempty"`
	TakerCommission fixedpoint.Value `json:"takerCommission,omitempty"`

	MakerFeeRate fixedpoint.Value `json:"makerFeeRate,omitempty"`
	TakerFeeRate fixedpoint.Value `json:"takerFeeRate,omitempty"`
	AccountType  string           `json:"accountType,omitempty"`

	TotalAccountValue fixedpoint.Value `json:"totalAccountValue,omitempty"`

	balances BalanceMap
}
