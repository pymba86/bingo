package engine

import (
	"context"
	"fmt"
	"github.com/pymba86/bingo/pkg/cmdutil"
	"github.com/pymba86/bingo/pkg/fixedpoint"
	"github.com/pymba86/bingo/pkg/types"
	log "github.com/sirupsen/logrus"
	"time"
)

type StandardIndicatorSet struct {
	Symbol string
	// Standard indicators
	// interval -> window
	store *MarketDataStore
}

func NewStandardIndicatorSet(symbol string, store *MarketDataStore) *StandardIndicatorSet {
	set := &StandardIndicatorSet{
		Symbol: symbol,
		store:  store,
	}

	return set
}

type ExchangeSession struct {
	Notifiability `json:"-" yaml:"-"`

	// ---------------------------
	// Session config fields
	// ---------------------------

	// Exchange Session name
	Name         string             `json:"name,omitempty" yaml:"name,omitempty"`
	ExchangeName types.ExchangeName `json:"exchange" yaml:"exchange"`
	EnvVarPrefix string             `json:"envVarPrefix" yaml:"envVarPrefix"`
	Key          string             `json:"key,omitempty" yaml:"key,omitempty"`
	Secret       string             `json:"secret,omitempty" yaml:"secret,omitempty"`
	SubAccount   string             `json:"subAccount,omitempty" yaml:"subAccount,omitempty"`

	// Withdrawal is used for enabling withdrawal functions
	Withdrawal   bool             `json:"withdrawal,omitempty" yaml:"withdrawal,omitempty"`
	MakerFeeRate fixedpoint.Value `json:"makerFeeRate,omitempty" yaml:"makerFeeRate,omitempty"`
	TakerFeeRate fixedpoint.Value `json:"takerFeeRate,omitempty" yaml:"takerFeeRate,omitempty"`

	PublicOnly           bool   `json:"publicOnly,omitempty" yaml:"publicOnly"`
	Margin               bool   `json:"margin,omitempty" yaml:"margin"`
	IsolatedMargin       bool   `json:"isolatedMargin,omitempty" yaml:"isolatedMargin,omitempty"`
	IsolatedMarginSymbol string `json:"isolatedMarginSymbol,omitempty" yaml:"isolatedMarginSymbol,omitempty"`

	// ---------------------------
	// Runtime fields
	// ---------------------------

	// The exchange account states
	Account *types.Account `json:"-" yaml:"-"`

	IsInitialized bool `json:"-" yaml:"-"`

	OrderExecutor *ExchangeOrderExecutor `json:"orderExecutor,omitempty" yaml:"orderExecutor,omitempty"`

	// UserDataStream is the connection stream of the exchange
	UserDataStream   types.Stream `json:"-" yaml:"-"`
	MarketDataStream types.Stream `json:"-" yaml:"-"`

	Subscriptions map[types.Subscription]types.Subscription `json:"-" yaml:"-"`

	Exchange types.Exchange `json:"-" yaml:"-"`

	// Trades collects the executed trades from the exchange
	// map: symbol -> []trade
	Trades map[string]*types.TradeSlice `json:"-" yaml:"-"`

	// markets defines market configuration of a symbol
	markets map[string]types.Market

	// orderBooks stores the streaming order book
	orderBooks map[string]*types.StreamOrderBook

	// startPrices is used for backtest
	startPrices map[string]float64

	lastPrices         map[string]float64
	lastPriceUpdatedAt time.Time

	// marketDataStores contains the market data store of each market
	marketDataStores map[string]*MarketDataStore

	positions map[string]*Position

	// standard indicators of each market
	standardIndicatorSets map[string]*StandardIndicatorSet

	orderStores map[string]*OrderStore

	usedSymbols        map[string]struct{}
	initializedSymbols map[string]struct{}

	logger *log.Entry

}

func InitExchangeSession(name string, session *ExchangeSession) error {
	var err error
	var exchangeName = session.ExchangeName
	var exchange types.Exchange
	if session.Key != "" && session.Secret != "" {
		if !session.PublicOnly {
			if len(session.Key) == 0 || len(session.Secret) == 0 {
				return fmt.Errorf("can not create exchange %s: empty key or secret", exchangeName)
			}
		}

		exchange, err = cmdutil.NewExchangeStandard(exchangeName, session.Key, session.Secret)
	} else {
		exchange, err = cmdutil.NewExchangeWithEnvVarPrefix(exchangeName, session.EnvVarPrefix)
	}

	if err != nil {
		return err
	}

	// configure exchange
	if session.Margin {
		marginExchange, ok := exchange.(types.MarginExchange)
		if !ok {
			return fmt.Errorf("exchange %s does not support margin", exchangeName)
		}

		if session.IsolatedMargin {
			marginExchange.UseIsolatedMargin(session.IsolatedMarginSymbol)
		} else {
			marginExchange.UseMargin()
		}
	}

	session.Name = name
	session.Notifiability = Notifiability{
		SymbolChannelRouter:  NewPatternChannelRouter(nil),
		SessionChannelRouter: NewPatternChannelRouter(nil),
		ObjectChannelRouter:  NewObjectChannelRouter(),
	}
	session.Exchange = exchange
	session.UserDataStream = exchange.NewStream()
	session.MarketDataStream = exchange.NewStream()
	session.MarketDataStream.SetPublicOnly()

	// pointer fields
	session.Subscriptions = make(map[types.Subscription]types.Subscription)
	session.Account = &types.Account{}
	session.Trades = make(map[string]*types.TradeSlice)

	session.orderBooks = make(map[string]*types.StreamOrderBook)
	session.markets = make(map[string]types.Market)
	session.lastPrices = make(map[string]float64)
	session.startPrices = make(map[string]float64)
	session.marketDataStores = make(map[string]*MarketDataStore)
	session.positions = make(map[string]*Position)
	session.standardIndicatorSets = make(map[string]*StandardIndicatorSet)
	session.orderStores = make(map[string]*OrderStore)
	session.OrderExecutor = &ExchangeOrderExecutor{
		// copy the notification system so that we can route
		Notifiability: session.Notifiability,
		Session:       session,
	}

	session.usedSymbols = make(map[string]struct{})
	session.initializedSymbols = make(map[string]struct{})
	session.logger = log.WithField("session", name)
	return nil
}

func (session *ExchangeSession) Init(ctx context.Context, environ *Environment) error {
	if session.IsInitialized {
		return ErrSessionAlreadyInitialized
	}

	var log = log.WithField("session", session.Name)

	markets, err := session.Exchange.QueryMarkets(ctx)
	if err != nil {
		return err
	}

	session.markets = markets

	// query and initialize the balances
	log.Infof("querying balances from session %s...", session.Name)
	balances, err := session.Exchange.QueryAccountBalances(ctx)
	if err != nil {
		return err
	}

	log.Infof("%s account", session.Name)
	balances.Print()

	session.Account.UpdateBalances(balances)

	// forward trade updates and order updates to the order executor
	session.UserDataStream.OnTradeUpdate(session.OrderExecutor.EmitTradeUpdate)
	session.UserDataStream.OnOrderUpdate(session.OrderExecutor.EmitOrderUpdate)
	session.Account.BindStream(session.UserDataStream)

	session.MarketDataStream.OnKLineClosed(func(kline types.KLine) {
		log.WithField("marketData", "kline").Infof("kline closed: %+v", kline)
	})

	// update last prices
	session.MarketDataStream.OnKLineClosed(func(kline types.KLine) {
		if _, ok := session.startPrices[kline.Symbol]; !ok {
			session.startPrices[kline.Symbol] = kline.Open
		}

		session.lastPrices[kline.Symbol] = kline.Close
	})

	session.IsInitialized = true
	return nil
}

