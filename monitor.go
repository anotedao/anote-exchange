package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/anonutopia/gowaves"
)

type Monitor struct {
	StartedTime int64
}

func (m *Monitor) start() {
	m.StartedTime = time.Now().Unix() * 1000
	var trades []*Trade

	for {
		db.Find(&trades)

		for _, trade := range trades {
			transactions := getTransactions(trade.AddressExchange)
			for _, t := range transactions {
				tjson, err := json.Marshal(t)
				if err != nil {
					log.Println(err)
				}

				tt := &TransferTransaction{}
				err = json.Unmarshal(tjson, tt)
				if err != nil {
					log.Println(err)
				}

				m.checkTransaction(tt, trade)
			}
		}

		time.Sleep(time.Second * MonitorTick)
	}
}

func (m *Monitor) checkTransaction(t *TransferTransaction, trade *Trade) {
	tr := &Transaction{TxID: t.ID}
	db.FirstOrCreate(tr, tr)

	if !tr.Processed {
		m.processTransaction(tr, t, trade)
	}
}

func (m *Monitor) processTransaction(tr *Transaction, t *TransferTransaction, trade *Trade) {
	if t.Type == 4 && t.Recipient == trade.AddressExchange {
		m.purchaseAsset(t, trade)
	}

	tr.Processed = true
	db.Save(tr)
}

func (m *Monitor) purchaseAsset(t *TransferTransaction, trade *Trade) {
	waves := t.Amount - 2*WavesFee - WavesExchangeFee

	if waves > 0 {
		a, p := m.calculateAssetAmount(uint64(waves))
		abr, err := gowaves.WNC.AddressesBalance(trade.AddressExchange)
		if err == nil {
			nabr, _ := gowaves.WNC.AddressesBalance(trade.AddressExchange)
			if purchaseAsset(a, uint64(waves), TokenID, p, trade.Seed) == nil {
				for abr.Balance == nabr.Balance {
					time.Sleep(time.Second * 10)
					nabr, _ = gowaves.WNC.AddressesBalance(trade.AddressExchange)
				}

				sendAsset(a, TokenID, trade.AddressUser, trade.Seed)
			}
		}
	}
}

func (m *Monitor) calculateAssetAmount(wavesAmount uint64) (amount uint64, price uint64) {
	opr, err := gowaves.WMC.OrderbookPair(TokenID, "WAVES", 10)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return 0, 0
	}

	waves := uint64(0)

	for _, a := range opr.Asks {
		if wavesAmount > 0 {
			w := a.Amount * a.Price / MULTI8
			newWaves := uint64(0)
			if w < wavesAmount {
				newWaves = w
				amount += a.Amount
				waves += newWaves
				wavesAmount -= newWaves
			} else {
				newWaves = wavesAmount
				amount += uint64(float64(wavesAmount) / float64(a.Price) * float64(MULTI8))
				waves += newWaves
				wavesAmount -= newWaves
			}
			price = a.Price
		}
	}

	return amount, price
}

func initMonitor() {
	m := &Monitor{}
	go m.start()
}

type TransferTransaction struct {
	Type            int         `json:"type"`
	Version         int         `json:"version"`
	ID              string      `json:"id"`
	Proofs          []string    `json:"proofs"`
	SenderPublicKey string      `json:"senderPublicKey"`
	AssetID         interface{} `json:"assetId"`
	FeeAssetID      interface{} `json:"feeAssetId"`
	Timestamp       int64       `json:"timestamp"`
	Amount          int         `json:"amount"`
	Fee             int         `json:"fee"`
	Recipient       string      `json:"recipient"`
}
