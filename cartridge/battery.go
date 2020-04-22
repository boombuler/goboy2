package cartridge

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

type Battery interface {
	Open() BatteryStream
	HasData() bool
}

type BatteryStream interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
}

type BatteryFactory func() Battery

type BatteryFile string

func GetBattery(romFileName string) BatteryFile {
	ext := filepath.Ext(romFileName)
	baseFileName := romFileName[:len(romFileName)-len(ext)]
	return BatteryFile(baseFileName + ".ram")
}

func (bf BatteryFile) HasData() bool {
	_, err := os.Stat(string(bf))
	return !os.IsNotExist(err)
}

func (bf BatteryFile) Open() BatteryStream {
	fn := string(bf)
	if _, err := os.Stat(fn); !os.IsNotExist(err) {
		ramFile, err := os.OpenFile(fn, os.O_RDWR, os.ModeExclusive)
		if err != nil {
			log.Fatal(err)
		}
		return ramFile
	} else {
		ramFile, err := os.Create(fn)
		if err != nil {
			log.Fatal(err)
		}
		return ramFile
	}
}
