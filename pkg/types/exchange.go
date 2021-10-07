package types

type ExchangeName string

const (
	ExchangeBinance = ExchangeName("binance")
)

type Exchange interface {
	Name() ExchangeName
}
