package main

import (
	"log"

	"gopkg.in/macaron.v1"
	"gorm.io/gorm"
)

var m *macaron.Macaron

var db *gorm.DB

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	m = initMacaron()

	db = initDb()

	initMonitor()

	m.Run("127.0.0.1", Port)
}
