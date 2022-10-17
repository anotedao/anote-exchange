package main

import (
	"log"
	"strconv"

	"gopkg.in/macaron.v1"
)

func calculateView(ctx *macaron.Context) {
	cr := &CalculateResponse{}

	amountStr := ctx.Params("amount")
	from := ctx.Params("from")
	to := ctx.Params("to")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		log.Println(err)
		cr.Error = err.Error()
	}

	cr.ResultInstant = calculateInstant(amount, from, to)
	cr.ResultDelay = calculateDelay(amount, from, to)

	ctx.Resp.Header().Add("Access-Control-Allow-Origin", "*")
	ctx.JSON(200, cr)
}

func tradeView(ctx *macaron.Context) {
	tr := &TradeResponse{}

	amountStr := ctx.Params("amount")
	from := ctx.Params("from")
	to := ctx.Params("to")
	ttype := ctx.Params("type")
	address := ctx.Params("address")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		log.Println(err)
		tr.Error = err.Error()
	} else {
		amountInt := uint64(amount * MULTI8)
		t := newTrade(from, to, amountInt, ttype, address)
		tr.Success = true
		tr.Address = t.AddressExchange
	}

	ctx.Resp.Header().Add("Access-Control-Allow-Origin", "*")
	ctx.JSON(200, tr)
}
