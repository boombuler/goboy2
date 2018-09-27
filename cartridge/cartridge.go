package cartridge

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type MBC interface {
	Read(addr uint16) byte
	Write(addr uint16, value byte)
	Shutdown()
}

type Cartridge struct {
	MBC
	Title    string
	GBC      bool
	SGB      bool
	ROMSize  uint
	RAMSize  uint
	Japanese bool
	Version  byte
}

var mbcFactories = map[byte]func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error){
	0x00: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// ROM Only
		return createMBC0(c, data, false, nil)
	},
	0x01: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// MBC1
		return createMBC1(c, data, nil)
	},
	0x02: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// MBC1+RAM
		return createMBC1(c, data, nil)
	},
	0x03: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// MBC1+RAM+BAT
		return createMBC1(c, data, bf())
	},
	0x05: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// MBC2
		return createMBC2(c, data, nil)
	},
	0x06: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// MBC2+BAT
		return createMBC2(c, data, bf())
	},
	0x08: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// ROM+RAM
		return createMBC0(c, data, true, nil)
	},
	0x09: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// ROM+RAM+BAT
		return createMBC0(c, data, true, bf())
	},
	// 0x0B MMM01
	// 0x0C MMM01+RAM
	// 0x0D MMM01+RAM+BAT
	0x0F: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// MBC3+Timer+BAT
		return createMBC3(c, data, true, bf())
	},
	0x10: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// MBC3+Timer+RAM+BAT
		return createMBC3(c, data, true, bf())
	},
	0x11: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// MBC3
		return createMBC3(c, data, false, nil)
	},
	0x12: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// MBC3+RAM
		return createMBC3(c, data, false, nil)
	},
	0x13: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// MBC3+RAM+BAT
		return createMBC3(c, data, false, bf())
	},
	// 0x15 MBC4
	// 0x16 MBC4+RAM
	// 0x17 MBC4+RAM+BAT
	0x19: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// 0x19 MBC5
		return createMBC5(c, data, nil)
	},
	0x1A: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// 0x1A MBC5+RAM
		return createMBC5(c, data, nil)
	},
	0x1B: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// 0x1B MBC5+RAM+BAT
		return createMBC5(c, data, bf())
	},
	0x1C: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// 0x1C MBC5+RUMBLE
		return createMBC5(c, data, nil)
	},
	0x1D: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// 0x1D MBC5+RUMBLE+RAM
		return createMBC5(c, data, nil)
	},
	0x1E: func(c *Cartridge, data []byte, bf BatteryFactory) (MBC, error) {
		// 0x1E MBC5+RUMBLE+RAM+BAT
		return createMBC5(c, data, bf())
	},
	// 0xFC POCKET CAMERA
	// 0xFD BANDAI TAMA5
	// 0xFE HuC3
	// 0xFF HuC1+RAM+BAT
}

func Load(reader io.Reader, bf BatteryFactory) (*Cartridge, error) {
	rom, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if len(rom) < 0x8000 {
		return nil, fmt.Errorf("Invalid ROM")
	}
	c := new(Cartridge)
	c.Title = strings.TrimRight(string(rom[0x0134:0x0142]), "\x00")
	c.GBC = (rom[0x0143] == 0x80) || (rom[0x0143] == 0xC0)
	c.SGB = rom[0x0146] == 0x03 && rom[0x014B] == 0x33
	c.ROMSize = 0x8000 << rom[0x0148]

	switch rom[0x0149] {
	case 0x00:
		c.RAMSize = 0x000000
	case 0x01:
		c.RAMSize = 0x000800
	case 0x02:
		c.RAMSize = 0x002000
	case 0x03:
		c.RAMSize = 0x008000
	case 0x04:
		c.RAMSize = 0x020000
	default:
		return nil, fmt.Errorf("Unsupported RAM size: %v", rom[0x0149])
	}
	c.Japanese = (rom[0x014A] == 0x00)
	c.Version = rom[0x014C]
	mbcFactory, ok := mbcFactories[rom[0x0147]]
	if !ok {
		return nil, fmt.Errorf("MBC type not supported: %02x", rom[0x0147])
	} else {
		c.MBC, err = mbcFactory(c, rom, bf)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}
