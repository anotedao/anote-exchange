package main

import (
	"log"

	"gopkg.in/macaron.v1"
)

func calculateView(ctx *macaron.Context) {
	log.Println("fdsa")
	ctx.JSON(200, nil)
}
