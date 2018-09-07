package cpu

func (c *CPU) incByte(value byte) byte {
	result := value + 0x01
	c.setFlag(substract, false)
	c.setFlag(zero, result == 0x00)
	c.setFlag(halfcarry, (result^value)&0x10 == 0x10)
	return result
}

func (c *CPU) decByte(value byte) byte {
	result := value - 1
	c.setFlag(substract, true)
	c.setFlag(zero, result == 0x00)
	c.setFlag(halfcarry, (result^value)&0x10 == 0x10)
	return result
}

func (c *CPU) addBytes(a, b byte) byte {
	result := a + b
	c.setFlag(substract, false)
	c.setFlag(zero, result == 0x00)
	c.setFlag(halfcarry, (result^b^a)&0x10 == 0x10)
	c.setFlag(carry, result < a)
	return result
}

func (c *CPU) subBytes(a, b byte) byte {
	result := a - b
	c.setFlag(substract, true)
	c.setFlag(zero, result == 0x00)
	c.setFlag(halfcarry, (result^b^a)&0x10 == 0x10)
	c.setFlag(carry, a < b)
	return result
}

func (c *CPU) sbcBytes(a, b byte) byte {
	orgA := a
	result := int(a) - int(b)
	if c.hasFlag(carry) {
		result--
	}
	a = byte(result)
	c.setFlag(carry, result < 0x00)
	c.setFlag(substract, true)
	c.setFlag(zero, a == 0x00)
	c.setFlag(halfcarry, ((a^b^orgA)&0x10) == 0x10)
	return a
}

func (c *CPU) addWords(a, b uint16) uint16 {
	result := a + b
	c.setFlag(carry, result < a)
	c.setFlag(halfcarry, (result^b^a)&0x1000 == 0x1000)
	c.setFlag(substract, false)
	return result
}

func (c *CPU) adcBytes(a, b byte) byte {
	cry := byte((c.f & carry) >> 4)
	c.setFlag(substract, false)
	c.setFlag(halfcarry, int(a&0x0F)+int(cry&0x0F)+int(b&0x0F) > 0xF)
	c.setFlag(carry, int(a)+int(cry)+int(b) > 0xFF)
	a += b + cry
	c.setFlag(zero, a == 0x00)
	return a
}

func (c *CPU) andBytes(a, b byte) byte {
	a = a & b
	c.setFlag(substract|carry, false)
	c.setFlag(halfcarry, true)
	c.setFlag(zero, a == 0x00)
	return a
}

func (c *CPU) xorBytes(a, b byte) byte {
	a = a ^ b
	c.setFlag(substract|carry|halfcarry, false)
	c.setFlag(zero, a == 0x00)
	return a
}

func (c *CPU) orBytes(a, b byte) byte {
	a = a | b
	c.setFlag(substract|carry|halfcarry, false)
	c.setFlag(zero, a == 0x00)
	return a
}

func (c *CPU) rlc(reg byte, checkZero bool) byte {
	hbit := reg&0x80 == 0x80
	result := reg << 1
	if hbit {
		c.setFlag(carry, true)
		result ^= 0x01
	} else {
		c.setFlag(carry, false)
	}
	c.setFlag(zero, checkZero && result == 0x00)
	c.setFlag(substract|halfcarry, false)
	return result
}

func (c *CPU) sla(reg byte) byte {
	c.setFlag(carry, reg&0x80 == 0x80)
	reg = reg << 1
	c.setFlag(zero, reg == 0x00)
	c.setFlag(substract|halfcarry, false)
	return reg
}

func (c *CPU) srl(reg byte) byte {
	c.setFlag(carry, reg&0x01 == 0x01)
	reg = (reg >> 1) & 0x7F
	c.setFlag(substract|halfcarry, false)
	c.setFlag(zero, reg == 0x00)
	return reg
}

func (c *CPU) sra(reg byte) byte {
	c.setFlag(carry, reg&0x01 == 0x01)
	reg = (reg >> 1) | (reg & 0x80)
	c.setFlag(substract|halfcarry, false)
	c.setFlag(zero, reg == 0x00)
	return reg
}

func (c *CPU) swap(reg byte) byte {
	reg = ((reg >> 4) & 0x0F) | ((reg << 4) & 0xF0)
	c.setFlag(substract|halfcarry|carry, false)
	c.setFlag(zero, reg == 0x00)
	return reg
}

func (c *CPU) testBit(num byte, reg byte) {
	c.setFlag(zero, (reg>>num)&0x01 == 0x00)
	c.setFlag(substract, false)
	c.setFlag(halfcarry, true)
}

func (c *CPU) setBit(num byte, reg byte) byte {
	return reg | (0x01 << num)
}

func (c *CPU) resetBit(num byte, reg byte) byte {
	return reg & ^(0x01 << num)
}

func (c *CPU) rrc(reg byte, checkZero bool) byte {
	lbit := reg&0x01 == 0x01
	result := reg >> 1
	if lbit {
		c.setFlag(carry, true)
		result ^= 0x80
	} else {
		c.setFlag(carry, false)
	}
	c.setFlag(zero, checkZero && result == 0x00)
	c.setFlag(substract|halfcarry, false)
	return result
}

func (c *CPU) rl(reg byte, checkZero bool) byte {
	hbit := reg&0x80 == 0x80
	reg = reg << 1

	if c.hasFlag(carry) {
		reg ^= 0x01
	}

	c.setFlag(carry, hbit)
	c.setFlag(zero, checkZero && reg == 0x00)
	c.setFlag(substract|halfcarry, false)
	return reg
}

func (c *CPU) rr(reg byte, checkZero bool) byte {
	lbit := reg&0x01 == 0x01
	reg = reg >> 1
	if c.hasFlag(carry) {
		reg ^= 0x80
	}
	c.setFlag(carry, lbit)
	c.setFlag(zero, checkZero && reg == 0x00)
	c.setFlag(substract|halfcarry, false)
	return reg
}

// Decimal adjust register A.
// This instruction adjusts register A so that the
// correct representation of Binary Coded Decimal (BCD)
// is obtained.
func (c *CPU) daa() {
	val := int(c.a)

	//Add or subtract correction values based on Subtract Flag
	if !c.hasFlag(substract) {
		if c.hasFlag(halfcarry) || (val&0x0F) > 0x09 {
			val += 0x06
		}
		if c.hasFlag(carry) || val > 0x9F {
			val += 0x60
		}
	} else {
		if c.hasFlag(halfcarry) {
			val = (val - 0x06) & 0xFF
		}
		if c.hasFlag(carry) {
			val = val - 0x60
		}
	}
	if (val & 0x0100) == 0x0100 {
		c.setFlag(carry, true)
	}
	val = val & 0xFF
	c.setFlag(halfcarry, false)
	c.setFlag(zero, val == 0)
	c.a = byte(val)
}
