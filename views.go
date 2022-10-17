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

	ctx.JSON(200, cr)
}
