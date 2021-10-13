package types

import "context"

type Channel string

type Stream interface {
	Subscribe(channel Channel, symbol string, options SubscribeOptions)
	SetPublicOnly()
	Connect(ctx context.Context) error
	Close() error
}

type SubscribeOptions struct {
	Interval string `json:"interval,omitempty"`
	Depth    string `json:"depth,omitempty"`
}

func (o SubscribeOptions) String() string {
	if len(o.Interval) > 0 {
		return o.Interval
	}

	return o.Depth
}

type Subscription struct {
	Symbol string `json:"symbol"`
	Channel Channel `json:"channel"`
	Options SubscribeOptions `json:"options"`
}
