package pnl

import (
	"github.com/pymba86/bingo/pkg/types"
	log "github.com/sirupsen/logrus"
	"time"
)

type AverageCostPnlReport struct {
	CurrentPrice float64
	StartTime    time.Time
	Symbol       string
	Market       types.Market

	NumTrades        int
	Profit           float64
	UnrealizedProfit float64
	AverageBidCost   float64
	BuyVolume        float64
	SellVolume       float64
	FeeInUSD         float64
	Stock            float64
	CurrencyFees     map[string]float64
}

func (report AverageCostPnlReport) Print() {
	log.Infof("TRADES SINCE: %v", report.StartTime)
	log.Infof("NUMBER OF TRADES: %d", report.NumTrades)
	log.Infof("AVERAGE COST: %s", types.USD.FormatMoneyFloat64(report.AverageBidCost))
	log.Infof("TOTAL BUY VOLUME: %f", report.BuyVolume)
	log.Infof("TOTAL SELL VOLUME: %f", report.SellVolume)
	log.Infof("STOCK: %f", report.Stock)
	log.Infof("FEE (USD): %f", report.FeeInUSD)
	log.Infof("CURRENT PRICE: %s", types.USD.FormatMoneyFloat64(report.CurrentPrice))
	log.Infof("CURRENCY FEES:")
	for currency, fee := range report.CurrencyFees {
		log.Infof(" - %s: %f", currency, fee)
	}
	log.Infof("PROFIT: %s", types.USD.FormatMoneyFloat64(report.Profit))
	log.Infof("UNREALIZED PROFIT: %s", types.USD.FormatMoneyFloat64(report.UnrealizedProfit))
}