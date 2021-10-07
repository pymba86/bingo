package types

type ExchangeName string

type Exchange interface {
	Name() ExchangeName
}
