package types

type Market struct {
	Symbol      string
	LocalSymbol string

	PricePrecision  int
	VolumePrecision int
	QuoteCurrency   string
	BaseCurrency    string

	// The MIN_NOTIONAL filter defines the minimum notional value allowed for an order on a symbol.
	// An order's notional value is the price * quantity
	MinNotional float64
	MinAmount   float64

	// The LOT_SIZE filter defines the quantity
	MinQuantity float64
	MaxQuantity float64
	StepSize    float64

	MinPrice float64
	MaxPrice float64
	TickSize float64
}