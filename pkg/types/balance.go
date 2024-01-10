package types

import (
	"fmt"
	"github.com/pymba86/bingo/pkg/fixedpoint"
	log "github.com/sirupsen/logrus"
	"strings"
)

type PriceMap map[string]fixedpoint.Value

type Balance struct {
	Currency  string           `json:"currency"`
	Available fixedpoint.Value `json:"available"`
	Locked    fixedpoint.Value `json:"locked,omitempty"`

	// margin related fields
	Borrowed fixedpoint.Value `json:"borrowed,omitempty"`
	Interest fixedpoint.Value `json:"interest,omitempty"`

	// NetAsset = (Available + Locked) - Borrowed - Interest
	NetAsset fixedpoint.Value `json:"net,omitempty"`

	MaxWithdrawAmount fixedpoint.Value `json:"maxWithdrawAmount,omitempty"`
}

func (b Balance) Add(b2 Balance) Balance {
	var newB = b
	newB.Available = b.Available.Add(b2.Available)
	newB.Locked = b.Locked.Add(b2.Locked)
	newB.Borrowed = b.Borrowed.Add(b2.Borrowed)
	newB.NetAsset = b.NetAsset.Add(b2.NetAsset)
	newB.Interest = b.Interest.Add(b2.Interest)
	return newB
}

func (b Balance) Total() fixedpoint.Value {
	return b.Available.Add(b.Locked)
}

func (b Balance) Net() fixedpoint.Value {
	total := b.Total()
	return total.Sub(b.Debt())
}

func (b Balance) Debt() fixedpoint.Value {
	return b.Borrowed.Add(b.Interest)
}

func (b Balance) ValueString() (o string) {
	o = b.Net().String()

	if b.Locked.Sign() > 0 {
		o += fmt.Sprintf(" (locked %v)", b.Locked)
	}

	if b.Borrowed.Sign() > 0 {
		o += fmt.Sprintf(" (borrowed: %v)", b.Borrowed)
	}

	return o
}

func (b Balance) String() (o string) {
	o = fmt.Sprintf("%s: %s", b.Currency, b.Net().String())

	if b.Locked.Sign() > 0 {
		o += fmt.Sprintf(" (locked %f)", b.Locked.Float64())
	}

	if b.Borrowed.Sign() > 0 {
		o += fmt.Sprintf(" (borrowed: %f)", b.Borrowed.Float64())
	}

	if b.Interest.Sign() > 0 {
		o += fmt.Sprintf(" (interest: %f)", b.Interest.Float64())
	}

	return o
}

type BalanceMap map[string]Balance

func (m BalanceMap) NotZero() BalanceMap {
	bm := make(BalanceMap)
	for c, b := range m {
		if b.Total().IsZero() && b.Debt().IsZero() && b.Net().IsZero() {
			continue
		}

		bm[c] = b
	}
	return bm
}

func (m BalanceMap) Debts() BalanceMap {
	bm := make(BalanceMap)
	for c, b := range m {
		if b.Borrowed.Sign() > 0 || b.Interest.Sign() > 0 {
			bm[c] = b
		}
	}
	return bm
}

func (m BalanceMap) Currencies() (currencies []string) {
	for _, b := range m {
		currencies = append(currencies, b.Currency)
	}
	return currencies
}

func (m BalanceMap) String() string {
	var ss []string
	for _, b := range m {
		ss = append(ss, b.String())
	}

	return "BalanceMap[" + strings.Join(ss, ", ") + "]"
}

func (m BalanceMap) Copy() (d BalanceMap) {
	d = make(BalanceMap)
	for c, b := range m {
		d[c] = b
	}
	return d
}

func (m BalanceMap) Print() {
	for _, balance := range m {
		if balance.Net().IsZero() {
			continue
		}

		o := fmt.Sprintf(" %s: %v", balance.Currency, balance.Available)
		if balance.Locked.Sign() > 0 {
			o += fmt.Sprintf(" (locked %v)", balance.Locked)
		}

		if balance.Borrowed.Sign() > 0 {
			o += fmt.Sprintf(" (borrowed %v)", balance.Borrowed)
		}

		log.Infoln(o)
	}
}
