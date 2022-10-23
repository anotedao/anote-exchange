package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/anonutopia/gowaves"
	"github.com/mr-tron/base58"
	wavesplatform "github.com/wavesplatform/go-lib-crypto"
	"github.com/wavesplatform/gowaves/pkg/client"
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

func calculateInstant(amount float64, from string) (float64, int) {
	obr := getOrderBook()
	amount2 := float64(0)
	amountInt := int64(amount) * MULTI8
	price := 0

	if from == "anote" {
		for _, bid := range obr.Bids {
			if amountInt >= bid.Amount {
				amountInt -= bid.Amount
				amount2 += float64(bid.Amount) / float64(MULTI8) * float64(bid.Price)
				price = bid.Price
			} else {
				amount2 += float64(amountInt) / float64(MULTI8) * float64(bid.Price)
				amountInt = 0
				price = bid.Price
			}
		}
	} else if from == "waves" {
		for _, ask := range obr.Asks {
			askAmount := ask.Amount * int64(ask.Price) / MULTI8
			if amountInt >= askAmount {
				amountInt -= askAmount
				amount2 += float64(ask.Amount)
				price = ask.Price
			} else {
				amount2 += float64(amountInt * MULTI8 / int64(ask.Price))
				amountInt = 0
				price = ask.Price
			}
		}
	}

	amount2 /= MULTI8

	return math.Floor(amount2*MULTI8) / MULTI8, price
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

func generateKeysAddress(seed string, from string) (public string, private string, address string) {
	c := wavesplatform.NewWavesCrypto()
	sd := wavesplatform.Seed(seed)
	pair := c.KeyPair(sd)

	pk := crypto.MustPublicKeyFromBase58(string(pair.PublicKey))
	a, err := proto.NewAddressFromPublicKey(proto.MainNetScheme, pk)
	if err != nil {
		log.Println(err.Error())
	}

	return string(pair.PublicKey), string(pair.PrivateKey), a.String()
}

func purchaseAsset(amountAsset uint64, amountWaves uint64, from string, price uint64, seed string) error {
	// if conf.Dev || conf.Debug {
	// 	return errors.New(fmt.Sprintf("Not purchasing asset (dev): %d - %d - %s - %d", amountAsset, amountWaves, assetId, price))
	// }

	var assetBytes []byte
	var orderType proto.OrderType

	pubKey, privKey, _ := generateKeysAddress(seed, "waves")

	// Create sender's public key from BASE58 string
	sender, err := crypto.NewPublicKeyFromBase58(pubKey)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	matcher, err := crypto.NewPublicKeyFromBase58(MatcherPublicKey)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	// Create sender's private key from BASE58 string
	sk, err := crypto.NewSecretKeyFromBase58(privKey)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	// Current time in milliseconds
	ts := time.Now().Unix() * 1000
	ets := time.Now().Add(time.Hour*24*29).Unix() * 1000

	assetBytes = crypto.MustBytesFromBase58(TokenID)

	asset, err := proto.NewOptionalAssetFromBytes(assetBytes)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	assetW, err := proto.NewOptionalAssetFromString("")
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	if from == "waves" {
		orderType = proto.Buy
	} else {
		orderType = proto.Sell
	}

	bo := proto.NewUnsignedOrderV1(sender, matcher, *asset, *assetW, orderType, price, amountAsset, uint64(ts), uint64(ets), WavesExchangeFee)

	err = bo.Sign(proto.MainNetScheme, sk)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	_, err = gowaves.WMC.OrderbookMarketAlt(bo)
	// _, err = gowaves.WMC.OrderbookAlt(bo)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	return nil
}

func getTransactions(address string) []proto.Transaction {
	baseUrl := ""

	if strings.HasPrefix(address, "3A") {
		baseUrl = AnoteNodeURL
	} else {
		baseUrl = WavesNodeURL
	}

	cl, err := client.NewClient(client.Options{BaseUrl: baseUrl, Client: &http.Client{}})
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		// return err
	}

	// Context to cancel the request execution on timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	addr := proto.MustAddressFromString(address)

	transactions, _, err := cl.Transactions.Address(ctx, addr, 10)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		// return err
	}

	return transactions
}

func sendAsset(amount uint64, assetId string, recipient string, seed string) error {
	var networkByte byte
	var nodeURL string
	var att proto.Attachment
	var err error

	pubKey, privKey, _ := generateKeysAddress(seed, "waves")

	if strings.HasPrefix(recipient, "3A") {
		att, err = proto.NewAttachmentFromBase58(base58.Encode([]byte(recipient)))
		if err != nil {
			log.Println(err)
		}
		recipient = GatewayWaves
		networkByte = proto.MainNetScheme
		nodeURL = WavesNodeURL
	}

	// Create sender's public key from BASE58 string
	sender, err := crypto.NewPublicKeyFromBase58(pubKey)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	// Create sender's private key from BASE58 string
	sk, err := crypto.NewSecretKeyFromBase58(privKey)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	// Current time in milliseconds
	ts := time.Now().Unix() * 1000

	asset, err := proto.NewOptionalAssetFromString(assetId)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	assetW, err := proto.NewOptionalAssetFromString("")
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	rec, err := proto.NewAddressFromString(recipient)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	tr := proto.NewUnsignedTransferWithSig(sender, *asset, *assetW, uint64(ts), amount, WavesFee, proto.Recipient{Address: &rec}, att)

	err = tr.Sign(networkByte, sk)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	// Create new HTTP client to send the transaction to public TestNet nodes
	client, err := client.NewClient(client.Options{BaseUrl: nodeURL, Client: &http.Client{}})
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	// Context to cancel the request execution on timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// // Send the transaction to the network
	_, err = client.Transactions.Broadcast(ctx, tr)
	if err != nil {
		log.Println(err)
		// logTelegram(err.Error())
		return err
	}

	return nil
}

func borrowFee(seed string) {
	networkByte := proto.MainNetScheme
	nodeURL := WavesNodeURL

	_, _, recipient := generateKeysAddress(seed, "waves")

	// Create sender's public key from BASE58 string
	sender, err := crypto.NewPublicKeyFromBase58(conf.PublicKey)
	if err != nil {
		log.Println(err)
	}

	// Create sender's private key from BASE58 string
	sk, err := crypto.NewSecretKeyFromBase58(conf.PrivateKey)
	if err != nil {
		log.Println(err)
	}

	// Current time in milliseconds
	ts := time.Now().Unix() * 1000

	assetW, err := proto.NewOptionalAssetFromString("")
	if err != nil {
		log.Println(err)
	}

	rec, err := proto.NewAddressFromString(recipient)
	if err != nil {
		log.Println(err)
	}

	amount := uint64(WavesExchangeFee)

	tr := proto.NewUnsignedTransferWithSig(sender, *assetW, *assetW, uint64(ts), amount, WavesFee, proto.Recipient{Address: &rec}, nil)

	err = tr.Sign(networkByte, sk)
	if err != nil {
		log.Println(err)
	}

	// Create new HTTP client to send the transaction to public TestNet nodes
	client, err := client.NewClient(client.Options{BaseUrl: nodeURL, Client: &http.Client{}})
	if err != nil {
		log.Println(err)
	}

	// Context to cancel the request execution on timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// // Send the transaction to the network
	_, err = client.Transactions.Broadcast(ctx, tr)
	if err != nil {
		log.Println(err)
	}
}

func returnFee(seed string) {
	networkByte := proto.MainNetScheme
	nodeURL := WavesNodeURL

	pubKey, privKey, _ := generateKeysAddress(seed, "waves")

	// Create sender's public key from BASE58 string
	sender, err := crypto.NewPublicKeyFromBase58(pubKey)
	if err != nil {
		log.Println(err)
	}

	// Create sender's private key from BASE58 string
	sk, err := crypto.NewSecretKeyFromBase58(privKey)
	if err != nil {
		log.Println(err)
	}

	// Current time in milliseconds
	ts := time.Now().Unix() * 1000

	assetW, err := proto.NewOptionalAssetFromString("")
	if err != nil {
		log.Println(err)
	}

	recPk, err := crypto.NewPublicKeyFromBase58(conf.PublicKey)
	if err != nil {
		log.Println(err)
	}

	rec, err := proto.NewAddressFromPublicKey(networkByte, recPk)
	if err != nil {
		log.Println(err)
	}

	amount := uint64(WavesExchangeFee)

	tr := proto.NewUnsignedTransferWithSig(sender, *assetW, *assetW, uint64(ts), amount, WavesFee, proto.Recipient{Address: &rec}, nil)

	err = tr.Sign(networkByte, sk)
	if err != nil {
		log.Println(err)
	}

	// Create new HTTP client to send the transaction to public TestNet nodes
	client, err := client.NewClient(client.Options{BaseUrl: nodeURL, Client: &http.Client{}})
	if err != nil {
		log.Println(err)
	}

	// Context to cancel the request execution on timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// // Send the transaction to the network
	_, err = client.Transactions.Broadcast(ctx, tr)
	if err != nil {
		log.Println(err)
	}
}
