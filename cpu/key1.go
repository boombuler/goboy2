package cpu

type key1Reg struct {
	dblSpeed bool //Bit 7: Current Speed     (0=Normal, 1=Double) (Read Only)
	prep     bool //Bit 0: Prepare Speed Switch (0=No, 1=Prepare) (Read/Write)
}

func (k *key1Reg) Read(addr uint16) byte {
	result := byte(0x7E)
	if k.dblSpeed {
		result |= 0x80
	}
	if k.prep {
		result |= 0x01
	}
	return result
}

func (k *key1Reg) Write(addr uint16, val byte) {
	k.prep = val&1 != 0
}

func (k *key1Reg) changeSpeed() bool {
	if k.prep {
		k.dblSpeed = !k.dblSpeed
		k.prep = false
		return true
	}
	return false
}
