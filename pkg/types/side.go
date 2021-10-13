package types

import (
	"encoding/json"
	"github.com/pkg/errors"
	"strings"
)

type SideType string

const (
	SideTypeBuy  = SideType("BUY")
	SideTypeSell = SideType("SELL")
	SideTypeSelf = SideType("SELF")

	// SideTypeBoth is only used for the configuration context
	SideTypeBoth = SideType("BOTH")
)

var ErrInvalidSideType = errors.New("invalid side type")

func StrToSideType(s string) (side SideType, err error) {
	switch strings.ToLower(s) {
	case "buy":
		side = SideTypeBuy

	case "sell":
		side = SideTypeSell

	case "both":
		side = SideTypeBoth

	default:
		err = ErrInvalidSideType
		return side, err

	}

	return side, err
}

func (side *SideType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	ss, err := StrToSideType(s)
	if err != nil {
		return err
	}

	*side = ss
	return nil
}

func (side SideType) Reverse() SideType {
	switch side {
	case SideTypeBuy:
		return SideTypeSell

	case SideTypeSell:
		return SideTypeBuy
	}

	return side
}

func (side SideType) String() string {
	return string(side)
}
