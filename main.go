package main

import (
	"log"

	"gopkg.in/macaron.v1"
)

var m *macaron.Macaron

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	m = initMacaron()

	m.Run("127.0.0.1", Port)
}
