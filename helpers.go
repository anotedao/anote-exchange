package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/mr-tron/base58"
	wavesplatform "github.com/wavesplatform/go-lib-crypto"
	"github.com/wavesplatform/gowaves/pkg/crypto"
	"github.com/wavesplatform/gowaves/pkg/proto"
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
	obr := getOrderBook()
	amount2 := float64(0)
	amountInt := int64(amount) * MULTI8

	if from == "anote" {
		for _, bid := range obr.Bids {
			if amountInt >= bid.Amount {
				amountInt -= bid.Amount
				amount2 += float64(bid.Amount) / float64(MULTI8) * float64(bid.Price)
			} else {
				amount2 += float64(amountInt) / float64(MULTI8) * float64(bid.Price)
				amountInt = 0
			}
		}
	} else if from == "waves" {
		for _, ask := range obr.Asks {
			askAmount := ask.Amount * int64(ask.Price) / MULTI8
			if amountInt >= askAmount {
				amountInt -= askAmount
				amount2 += float64(ask.Amount)
			} else {
				amount2 += float64(amountInt * MULTI8 / int64(ask.Price))
				amountInt = 0
			}
		}
	}

	amount2 /= MULTI8

	return math.Floor(amount2*MULTI8) / MULTI8
}

func calculateDelay(amount float64, from string, to string) float64 {
	obr := getOrderBook()

	if from == "anote" {
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

type TradeResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Address string `json:"address"`
}

func urlToLines(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return linesFromReader(resp.Body)
}

func linesFromReader(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func getRandNum() int {
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 2048
	rn := rand.Intn(max-min+1) + min
	return rn
}

func generateSeed() (seed string, encoded string) {
	var words []string
	seed = ""
	encoded = ""

	lines, err := urlToLines(SeedWordsURL)
	if err != nil {
		log.Println(err.Error())
	}

	for _, line := range lines {
		words = append(words, line)
	}

	for i := 1; i <= 15; i++ {
		seed += words[getRandNum()]
		if i < 15 {
			seed += " "
		}
	}

	data := []byte(seed)
	encoded = base58.Encode(data)

	return seed, encoded
}

func generateKeysAddress(seed string) (public string, private string, address string) {
	c := wavesplatform.NewWavesCrypto()
	sd := wavesplatform.Seed(seed)
	pair := c.KeyPair(sd)

	pk := crypto.MustPublicKeyFromBase58(string(pair.PublicKey))
	a, err := proto.NewAddressFromPublicKey(55, pk)
	if err != nil {
		log.Println(err.Error())
	}

	return string(pair.PublicKey), string(pair.PrivateKey), a.String()
}
