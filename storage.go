package main

import "math/bits"

const (
	ACC uint8 = 0
	IX  uint8 = 1

	DataMemoryOffset = 256
)

type Storage struct {
	ACC  uint8
	IX   uint8
	PC   uint8
	MAR  uint8
	IR   uint8
	FLAG uint8

	Memory []uint8

	halt       bool
	ibuf, obuf uint8

	obufFlag uint8
	ibufFlag uint8
}

func (s *Storage) Halt() {
	s.halt = true
}

func (s *Storage) set(index uint8, value bool) {
	s.FLAG = s.FLAG & bits.Reverse8((uint8(1) << index))
	if value {
		s.FLAG |= 1 << index
	}
}

func (s *Storage) setZF(value bool) {
	s.set(0, value)
}

func (s *Storage) setNF(value bool) {
	s.set(1, value)
}

func (s *Storage) setVF(value bool) {
	s.set(2, value)
}

func (s *Storage) setCF(value bool) {
	s.set(3, value)
}

func (s *Storage) get(index uint8) bool {
	return (s.FLAG>>index)&1 != 0
}

func (s *Storage) getZF() bool {
	return s.get(0)
}

func (s *Storage) getNF() bool {
	return s.get(1)
}

func (s *Storage) getVF() bool {
	return s.get(2)
}

func (s *Storage) getCF() bool {
	return s.get(3)
}

func (s *Storage) getRegister(a uint8) uint8 {
	if a == ACC {
		return s.ACC
	}
	return s.IX
}

func (s *Storage) setRegister(a, v uint8) {
	if a == 0 {
		s.ACC = v
	}
	s.IX = v
}

func NewStorage() *Storage {
	return &Storage{
		Memory: make([]uint8, 512),
	}
}
