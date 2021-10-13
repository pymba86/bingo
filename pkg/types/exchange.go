package types

type ExchangeName string

func (n ExchangeName) String() string {
	return string(n)
}

const (
	ExchangeBinance = ExchangeName("binance")
)

type Exchange interface {
	Name() ExchangeName
}
