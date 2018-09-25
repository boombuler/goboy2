package main

import (
	"flag"
	"goboy2/cartridge"
	"goboy2/screen"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
)

func showUsage() {
	log.Println("Usage:")
	log.Println(filepath.Base(os.Args[0]), "(romfile)")
	os.Exit(1)
}

func loadCatridge() (*cartridge.Cartridge, error) {
	if flag.NArg() != 1 {
		showUsage()
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return cartridge.Load(f)
}

func ramFileName() string {
	romFile := flag.Arg(0)
	ext := filepath.Ext(romFile)
	baseFileName := romFile[:len(romFile)-len(ext)]
	return baseFileName + ".ram"
}

var (
	dump       = flag.Bool("dump", false, "dump cpu state after every instruction")
	noboot     = flag.Bool("noboot", false, "skip boot sequence")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	mooneye    = flag.Bool("mooneye", false, "runs a mooneye test-rom")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	c, err := loadCatridge()
	if err != nil {
		log.Fatal(err)
	}

	ramFileName := ramFileName()
	if c.HasBattery() {
		if _, err := os.Stat(ramFileName); !os.IsNotExist(err) {
			ramFile, err := os.Open(ramFileName)
			if err == nil {
				c.LoadRAM(ramFile)
				ramFile.Close()
			}
		}
	}

	if *mooneye {
		runMooneyeRom(c)
		return
	}

	screen.Main(func(s *screen.Screen, input <-chan interface{}, exitChan <-chan struct{}) {
		gb := NewGameBoy(c, s.GetOutputChannel())
		go func() {
			for {
				select {
				case _, _ = <-exitChan:
					if c.HasBattery() {
						ramFile, err := os.Create(ramFileName)
						if err == nil {
							c.DumpRAM(ramFile)
							ramFile.Close()
						}
					}
					return
				case ev := <-input:
					switch e := ev.(type) {
					case screen.KeyEvent:
						gb.kb.HandleKeyEvent(e.Pressed, e.Key)
					}
				}
			}
		}()

		if *noboot {
			gb.SetupNoBootRom()
		}
		gb.CPU.Dump = *dump
		gb.Run(exitChan)
	})
}
