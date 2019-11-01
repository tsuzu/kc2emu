package main

type Storage struct {
	ACC  uint8
	IX   uint8
	PC   uint8
	MAR  uint8
	IR   uint8
	FLAG uint8

	Memory []uint8
}

func NewStorage() *Storage {
	return &Storage{
		Memory: make([]uint8, 512),
	}
}
