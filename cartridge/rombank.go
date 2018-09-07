package cartridge

const (
	rombankSize = 0x4000
	rambankSize = 0x2000
)

type rombank []byte

func (m rombank) Read(addr uint16) byte {
	return m[addr%rombankSize]
}

type rambank [rambankSize]byte
