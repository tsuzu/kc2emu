package emulator

type Phase func(*Storage) Phase

func p0(s *Storage) Phase {
	s.MAR = s.PC
	s.PC++

	return p1
}

func p1(s *Storage) Phase {
	s.IR = s.Memory[s.MAR]

	up := s.IR >> 4
	cc := s.IR & 0b1111
	a := cc >> 3
	b := s.IR & 0b111
	sr := (s.IR & 0b100) >> 2
	sm := s.IR & 0b11
	switch up {
	case 0b0000:
		if a == 1 {
			return haltP2
		}

		return nopP2
	case 0b0101:
		return haltP2 // not available
	case 0b0001:
		if a == 0 {
			return outP2
		}

		return inP2
	case 0b0010:
		return setCFP1(a)
	case 0b0011:
		return branchP1(cc)
	case 0b0100:
		return shiftRotateP1(a, sr, sm)
	case 0b0110:
		return loadP1(a, b)
	case 0b0111:
		return storeP1(a, b)
	case 0b1000, 0b1001, 0b1010, 0b1011, 0b1110, 0b1111:
		return aluP1(up, a, b)
	}

	panic("unknown operation")
}

func inRegisterMode(b uint8) bool {
	return b&0b110 == 0b000
}

func isImmediate(b uint8) bool {
	return (b >> 1) == 0b01
}

func hasIXOffset(b uint8) bool {
	return (b & 0b10) != 0
}

func inDataMemory(b uint8) bool {
	return (b & 1) != 0
}

func alu(s *Storage, up, left, right uint8) uint8 {
	if (up>>1) == 0b100 && s.getCF() {
		right += 1
	}

	tcf := false
	tvf := false
	switch up {
	case 0b1001, 0b1011: // ADC/ADD
		tcf = uint16(left+right) != uint16(left)+uint16(right)
		tvf = int16(int8(left))+int16(int8(right)) != int16(int8(left)+int8(right))

		left += right
	case 0b1000, 0b1010: // SBC/SUB
		tcf = left < right
		tvf = int16(int8(left))-int16(int8(right)) != int16(int8(left)-int8(right))

		left -= right
	case 0b1100: // EOR
		left ^= right
	case 0b1101: // OR
		left |= right
	case 0b1110: // AND
		left &= right
	case 0b1111: // CMP
		tcf = left > right
		tvf = int16(int8(left))-int16(int8(right)) != int16(int8(left)-int8(right))

		left -= right
	}

	// Set CF
	if (up >> 1) == 0b100 {
		s.setCF(tcf)
	}

	if up != 0b1100 && up != 0b1101 && up != 0b1110 {
		s.setVF(tvf)
	} else {
		s.setVF(false)
	}

	s.setNF(left&0x80 != 0)
	s.setZF(left == 0)

	return left
}

func aluP1(up, a, b uint8) Phase {
	assignRes := func() bool {
		return up != 0b1111
	}

	calc := func(s *Storage, right uint8) {
		v := alu(s, up, s.getRegister(a), right)

		if assignRes() {
			s.setRegister(a, v)
		}
	}

	regP2 := func(s *Storage) Phase {
		calc(s, s.getRegister(b))

		return nil
	}
	addrP4 := func(s *Storage) Phase {
		var v uint8

		if inDataMemory(b) {
			v = s.Memory[uint(s.MAR)+DataMemoryOffset]
		} else {
			v = s.Memory[s.MAR]
		}

		calc(s, v)

		return nil
	}

	addrP3 := func(s *Storage) Phase {
		if isImmediate(b) {
			calc(s, s.Memory[s.MAR])

			return nil
		}

		addr := s.Memory[s.MAR]

		if hasIXOffset(b) {
			addr += s.IX
		}

		s.MAR = addr

		return addrP4
	}

	addrP2 := func(s *Storage) Phase {
		s.MAR = s.PC
		s.PC++

		return addrP3
	}

	return func(s *Storage) Phase {
		if inRegisterMode(b) {
			return regP2(s)
		}

		return addrP2(s)
	}
}

func storeP1(a, b uint8) Phase {
	p4 := func(s *Storage) Phase {
		if inDataMemory(b) {
			s.Memory[int(s.MAR)+DataMemoryOffset] = s.getRegister(a)
		} else {
			s.Memory[s.MAR] = s.getRegister(a)
		}

		return nil
	}

	p3 := func(s *Storage) Phase {
		addr := s.Memory[s.MAR]

		if hasIXOffset(b) {
			addr += s.IX
		}

		s.MAR = addr

		return p4
	}

	if inRegisterMode(b) {
		panic("storing to register is not supported")
	}

	if isImmediate(b) {
		panic("storing to immediate value is not supported")
	}

	return func(s *Storage) Phase {
		s.MAR = s.PC
		s.PC++

		return p3
	}
}

func loadP1(a, b uint8) Phase {
	regP2 := func(s *Storage) Phase {
		s.setRegister(a, s.getRegister(b))

		return nil
	}

	addrP4 := func(s *Storage) Phase {
		var v uint8

		if inDataMemory(b) {
			v = s.Memory[uint(s.MAR)+DataMemoryOffset]
		} else {
			v = s.Memory[s.MAR]
		}

		s.setRegister(a, v)

		return nil
	}

	addrP3 := func(s *Storage) Phase {
		if isImmediate(b) {
			s.setRegister(a, s.Memory[s.MAR])

			return nil
		}

		addr := s.Memory[s.MAR]

		if hasIXOffset(b) {
			addr += s.IX
		}

		s.MAR = addr

		return addrP4
	}

	addrP2 := func(s *Storage) Phase {
		s.MAR = s.PC
		s.PC++

		return addrP3
	}

	return func(s *Storage) Phase {
		if inRegisterMode(b) {
			return regP2(s)
		}

		return addrP2(s)
	}
}

func shiftRotateP1(a, sr, sm uint8) Phase {
	p3 := func(v uint8) Phase {
		return func(s *Storage) Phase {
			s.setVF(false)
			s.setNF((v & 0x80) != 0)
			s.setZF(v == 0)
			return nil
		}
	}
	p3v := func(v uint8) Phase {
		return func(s *Storage) Phase {
			s.setVF(s.getCF() != (v&0x80 != 0))
			s.setNF((v & 0x80) != 0)
			s.setZF(v == 0)
			return nil
		}
	}

	sft := func(s *Storage, v uint8) (uint8, Phase) {
		switch sm {
		case 0b00: // SRA
			s.setCF(v&1 != 0)
			v = (v >> 1) | (v & 0x80)

			return v, p3(v)
		case 0b01: // SLA
			s.setCF((v & 0x80) != 0)
			v <<= 1

			return v, p3v(v)
		case 0b10: // SRL
			s.setCF(v&1 != 0)
			v >>= 1

			return v, p3(v)
		case 0b11: // SLL
			s.setCF((v & 0x80) != 0)
			v <<= 1

			return v, p3(v)
		}

		panic("invalid shift mode")
	}

	rot := func(s *Storage, v uint8) (uint8, Phase) {
		tcf := s.getCF()

		switch sm {
		case 0b00: // RRA
			s.setCF(v&1 != 0)
			v >>= 1
			if tcf {
				v |= 0x80
			}

			return v, p3(v)
		case 0b01: // RLA
			s.setCF((v & 0x80) != 0)
			v <<= 1

			if tcf {
				v |= 0x01
			}

			return v, p3v(v)
		case 0b10: // RRL
			s.setCF(v&1 != 0)
			v = (v >> 1) | ((v & 1) << 7)

			return v, p3(v)
		case 0b11: // RLL
			s.setCF((v & 0x80) != 0)
			v = (v << 1) | ((v & 0x80) >> 7)

			return v, p3(v)
		}

		panic("invalid rotate mode")
	}

	p2 := func(s *Storage) Phase {
		v := s.getRegister(a)
		var next Phase

		if sr == 0 { // S
			v, next = sft(s, v)
		} else {
			v, next = rot(s, v)
		}
		s.setRegister(a, v)

		return next
	}

	return p2
}

func setCFP1(v uint8) Phase {
	if v != 0 {
		v = 1
	}
	return func(s *Storage) Phase {
		s.setCF(v == 1)

		return nil
	}
}

func branchP1(cc uint8) Phase {
	checker := func(s *Storage) bool {
		switch cc {
		case 0b0000: // A
			return true
		case 0b1000: // VF
			return s.getVF()
		case 0b0001: // NZ
			return !s.getZF()
		case 0b1001: // Z
			return s.getZF()
		case 0b0010: // ZP
			return !s.getNF()
		case 0b1010: // N
			return s.getNF()
		case 0b0011: // P
			return !(s.getNF() || s.getZF())
		case 0b1011: // ZN
			return s.getNF() || s.getZF()
		case 0b0100: // NI
			return s.ibufFlag == 0
		case 0b1100: // NO
			return s.obufFlag == 1
		case 0b0101: // NC
			return !s.getCF()
		case 0b1101: // C
			return s.getCF()
		case 0b0110: // GE
			return !(!s.getVF() && s.getNF()) || (s.getVF() && !s.getNF()) // xor -> not
		case 0b1110: // LT
			return (!s.getVF() && s.getNF()) || (s.getVF() && !s.getNF()) // xor
		case 0b0111:
			xor := (!s.getVF() && s.getNF()) || (s.getVF() && !s.getNF())

			return !(xor || s.getZF())
		case 0b1111:
			xor := (!s.getVF() && s.getNF()) || (s.getVF() && !s.getNF())

			return (xor || s.getZF())
		default:
			panic("unknown branch condition code")
		}
	}

	p3 := func(s *Storage) Phase {
		if checker(s) {
			s.PC = s.Memory[s.MAR]
		}

		return nil
	}

	p2 := func(s *Storage) Phase {
		s.MAR = s.PC
		s.PC++

		return p3
	}

	return p2
}

func haltP2(s *Storage) Phase {
	s.Halt()

	return nil
}

func nopP2(s *Storage) Phase {
	// nop

	return nil
}

func outP2(s *Storage) Phase {
	s.obuf = s.ACC

	return func(s *Storage) Phase {
		s.obufFlag = 1

		obufKicker := s.obufKicker
		go func() {
			if obufKicker != nil {
				obufKicker()
			}
		}()

		return nil
	}
}

func inP2(s *Storage) Phase {
	s.ACC = s.ibuf

	return func(s *Storage) Phase {
		s.ibufFlag = 0

		ibufKicker := s.ibufKicker
		go func() {
			if ibufKicker != nil {
				ibufKicker()
			}
		}()

		return nil
	}
}
