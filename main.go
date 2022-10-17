package main

import (
	"log"

	"gopkg.in/macaron.v1"
)

var m *macaron.Macaron

func main() {
	m = initMacaron()
	log.Println("Test.")

	m.Run("127.0.0.1", Port)
}
