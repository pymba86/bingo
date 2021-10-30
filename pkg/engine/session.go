package engine

import (
	"context"
	"fmt"
	"github.com/pymba86/bingo/pkg/cmdutil"
	"github.com/pymba86/bingo/pkg/fixedpoint"
	"github.com/pymba86/bingo/pkg/service"
	"github.com/pymba86/bingo/pkg/types"
	"github.com/pymba86/bingo/pkg/util"
	log "github.com/sirupsen/logrus"
	"strings"
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

func (session *ExchangeSession) InitSymbols(ctx context.Context, environ *Environment) error {
	if err := session.initUsedSymbols(ctx, environ); err != nil {
		return err
	}

	return nil
}

func (session *ExchangeSession) initUsedSymbols(ctx context.Context, environ *Environment) error {
	for symbol := range session.usedSymbols {
		if err := session.initSymbol(ctx, environ, symbol); err != nil {
			return err
		}
	}

	return nil
}

func (session *ExchangeSession) initSymbol(ctx context.Context, environ *Environment, symbol string) error {

	if _, ok := session.initializedSymbols[symbol]; ok {
		return nil
	}

	market, ok := session.markets[symbol]
	if !ok {
		return fmt.Errorf("market %s is not defined", symbol)
	}

	var err error
	var trades []types.Trade
	if environ.SyncService != nil {
		tradingFeeCurrency := session.Exchange.PlatformFeeCurrency()
		if strings.HasPrefix(symbol, tradingFeeCurrency) {
			trades, err = environ.TradeService.QueryForTradingFeeCurrency(session.Exchange.Name(), symbol, tradingFeeCurrency)
		} else {
			trades, err = environ.TradeService.Query(service.QueryTradesOptions{
				Exchange: session.Exchange.Name(),
				Symbol:   symbol,
			})
		}

		if err != nil {
			return err
		}

		log.Infof("symbol %s: %d trades loaded", symbol, len(trades))
	}

	session.Trades[symbol] = &types.TradeSlice{Trades: trades}
	session.UserDataStream.OnTradeUpdate(func(trade types.Trade) {
		session.Trades[symbol].Append(trade)
	})

	position := &Position{
		Symbol:        symbol,
		BaseCurrency:  market.BaseCurrency,
		QuoteCurrency: market.QuoteCurrency,
	}
	position.AddTrades(trades)
	position.BindStream(session.UserDataStream)
	session.positions[symbol] = position

	session.initializedSymbols[symbol] = struct{}{}
	return nil
}

func (session *ExchangeSession) FindPossibleSymbols() (symbols []string, err error) {
	// If the session is an isolated margin session, there will be only the isolated margin symbol
	if session.Margin && session.IsolatedMargin {
		return []string{
			session.IsolatedMarginSymbol,
		}, nil
	}

	var balances = session.Account.Balances()
	var fiatAssets []string

	for _, currency := range types.FiatCurrencies {
		if balance, ok := balances[currency]; ok && balance.Total() > 0 {
			fiatAssets = append(fiatAssets, currency)
		}
	}

	var symbolMap = map[string]struct{}{}

	for _, market := range session.Markets() {
		// ignore the markets that are not fiat currency markets
		if !util.StringSliceContains(fiatAssets, market.QuoteCurrency) {
			continue
		}

		// ignore the asset that we don't have in the balance sheet
		balance, hasAsset := balances[market.BaseCurrency]
		if !hasAsset || balance.Total() == 0 {
			continue
		}

		symbolMap[market.Symbol] = struct{}{}
	}

	for s := range symbolMap {
		symbols = append(symbols, s)
	}

	return symbols, nil
}

func (session *ExchangeSession) Markets() map[string]types.Market {
	return session.markets
}
