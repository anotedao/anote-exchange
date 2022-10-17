package main

import (
	"gorm.io/gorm"
)

type Trade struct {
	gorm.Model
	Type            string `gorm:"size:255"`
	From            string `gorm:"size:255"`
	To              string `gorm:"size:255"`
	AddressUser     string `gorm:"size:255"`
	AddressExchange string `gorm:"size:255"`
	Seed            string `gorm:"size:255"`
	Amount          uint64 `gorm:"type:int"`
}

func newTrade(from string, to string, amount uint64, ttype string, address string) *Trade {
	t := &Trade{AddressUser: address}
	db.FirstOrCreate(t, t)

	if len(t.Type) == 0 {
		t.Amount = amount
		t.AddressUser = address
		t.Type = ttype
		t.Seed, _ = generateSeed()
		_, _, t.AddressExchange = generateKeysAddress(t.Seed)
		t.From = from
		t.To = to

		db.Save(t)
	}

	return t
}
