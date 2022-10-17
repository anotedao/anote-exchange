package main

import (
	"github.com/go-macaron/cache"
	macaron "gopkg.in/macaron.v1"
)

func initMacaron() *macaron.Macaron {
	m := macaron.Classic()

	m.Use(cache.Cacher())
	m.Use(macaron.Renderer())

	m.Get("/calculate/:from/:to/:amount", calculateView)
	m.Get("/trade/:from/:to/:amount/:type/:address", tradeView)

	return m
}
