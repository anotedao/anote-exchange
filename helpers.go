package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math"
	"net/http"
)

func getOrderBook() *OrderbookResponse {
	obr := &OrderbookResponse{}

	cl := http.Client{}

	var req *http.Request
	var err error

	req, err = http.NewRequest(http.MethodGet, OrderbookURL, nil)

	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return nil
	}

	res, err := cl.Do(req)

	if err == nil {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			// logTelegram(err.Error())
			return nil
		}
		if res.StatusCode != 200 {
			err := errors.New(res.Status)
			log.Println(err)
			// logTelegram(err.Error())
			return nil
		}
		json.Unmarshal(body, obr)
	} else {
		log.Println(err)
		// logTelegram(err.Error())
		return nil
	}

	return obr
}

type OrderbookResponse struct {
	Timestamp int64 `json:"timestamp"`
	Pair      struct {
		AmountAsset string `json:"amountAsset"`
		PriceAsset  string `json:"priceAsset"`
	} `json:"pair"`
	Bids []struct {
		Amount int64 `json:"amount"`
		Price  int   `json:"price"`
	} `json:"bids"`
	Asks []struct {
		Amount int64 `json:"amount"`
		Price  int   `json:"price"`
	} `json:"asks"`
}

type CalculateResponse struct {
	Success       bool    `json:"success"`
	Error         string  `json:"error"`
	ResultInstant float64 `json:"result_instant"`
	ResultDelay   float64 `json:"result_delay"`
}

func calculateInstant(amount float64, from string, to string) float64 {
	return 0
}

func calculateDelay(amount float64, from string, to string) float64 {
	obr := getOrderBook()

	if from == "anote" {
		log.Println(obr.Asks[0].Price - AskStep)
		amount2 := float64(obr.Asks[0].Price-AskStep) / float64(MULTI8) * amount
		return math.Floor(amount2*MULTI8) / MULTI8
	} else if from == "waves" {
		amount2 := amount / (float64(obr.Bids[0].Price+BidStep) / float64(MULTI8))
		return math.Floor(amount2*MULTI8) / MULTI8
	}

	return 0
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
