package engine

import "github.com/pymba86/bingo/pkg/types"

const MaxNumOfKLines = 5_000
const MaxNumOfKLinesTruncate = 1_000

// MarketDataStore receives and maintain the public market data
//go:generate callbackgen -type MarketDataStore
type MarketDataStore struct {
	Symbol string

	// KLineWindows stores all loaded klines per interval
	KLineWindows map[types.Interval]types.KLineWindow `json:"-"`

	kLineWindowUpdateCallbacks []func(interval types.Interval, kline types.KLineWindow)
}

func NewMarketDataStore(symbol string) *MarketDataStore {
	return &MarketDataStore{
		Symbol: symbol,

		// KLineWindows stores all loaded klines per interval
		KLineWindows: make(map[types.Interval]types.KLineWindow, len(types.SupportedIntervals)), // 12 interval, 1m,5m,15m,30m,1h,2h,4h,6h,12h,1d,3d,1w
	}
}

func (store *MarketDataStore) SetKLineWindows(windows map[types.Interval]types.KLineWindow) {
	store.KLineWindows = windows
}

// KLinesOfInterval returns the kline window of the given interval
func (store *MarketDataStore) KLinesOfInterval(interval types.Interval) (kLines types.KLineWindow, ok bool) {
	kLines, ok = store.KLineWindows[interval]
	return kLines, ok
}

func (store *MarketDataStore) BindStream(stream types.Stream) {
	stream.OnKLineClosed(store.handleKLineClosed)
}

func (store *MarketDataStore) handleKLineClosed(kline types.KLine) {
	if kline.Symbol != store.Symbol {
		return
	}

	store.AddKLine(kline)
}

func (store *MarketDataStore) AddKLine(kline types.KLine) {
	window, ok := store.KLineWindows[kline.Interval]
	if !ok {
		window = types.KLineWindow{kline}
	} else {
		window.Add(kline)
	}

	if len(window) > MaxNumOfKLines {
		window = window[MaxNumOfKLinesTruncate:]
	}

	store.KLineWindows[kline.Interval] = window
	store.EmitKLineWindowUpdate(kline.Interval, window)
}

func (store *MarketDataStore) OnKLineWindowUpdate(cb func(interval types.Interval, kline types.KLineWindow)) {
	store.kLineWindowUpdateCallbacks = append(store.kLineWindowUpdateCallbacks, cb)
}

func (store *MarketDataStore) EmitKLineWindowUpdate(interval types.Interval, kline types.KLineWindow) {
	for _, cb := range store.kLineWindowUpdateCallbacks {
		cb(interval, kline)
	}
}