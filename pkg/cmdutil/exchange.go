package cmdutil

import (
	"fmt"
	"github.com/pymba86/bingo/pkg/exchange/binance"
	"github.com/pymba86/bingo/pkg/types"
	"os"
	"strings"
)

func NewExchangeStandard(n types.ExchangeName, key, secret string) (types.Exchange, error) {
	switch n {
	case types.ExchangeBinance:
		return binance.New(key, secret), nil

	default:
		return nil, fmt.Errorf("unsupported exchange: %v", n)

	}
}

func NewExchangeWithEnvVarPrefix(n types.ExchangeName, varPrefix string) (types.Exchange, error) {
	if len(varPrefix) == 0 {
		varPrefix = n.String()
	}

	varPrefix = strings.ToUpper(varPrefix)

	key := os.Getenv(varPrefix + "_API_KEY")
	secret := os.Getenv(varPrefix + "_API_SECRET")
	if len(key) == 0 || len(secret) == 0 {
		return nil, fmt.Errorf("can not initialize exchange %s: empty key or secret, env var prefix: %s", n, varPrefix)
	}

	return NewExchangeStandard(n, key, secret)
}

func NewExchange(n types.ExchangeName) (types.Exchange, error) {
	return NewExchangeWithEnvVarPrefix(n, "")
}