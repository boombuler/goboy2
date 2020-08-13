package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"

	"github.com/boombuler/goboy2/consts"
	"github.com/boombuler/goboy2/mmu"

	"github.com/boombuler/goboy2/cartridge"
	"github.com/boombuler/goboy2/screen"

	"github.com/veandco/go-sdl2/sdl"
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

	bf := func() cartridge.Battery {
		return cartridge.GetBattery(flag.Arg(0))
	}

	return cartridge.Load(f, bf)
}

var (
	dump       = flag.Bool("dump", false, "dump cpu state after every instruction")
	noboot     = flag.Bool("noboot", false, "skip boot sequence")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	mooneye    = flag.Bool("mooneye", false, "runs a mooneye test-rom")
	gbc        = flag.Bool("color", false, "Force Gameboy Color mode")
	dmg        = flag.Bool("dmg", false, "Force DMG-Gameboy mode")
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

	hw := compatAuto
	if *gbc {
		hw = consts.GBC
	} else if *dmg {
		hw = consts.DMG
	}

	if *mooneye {
		runMooneyeRom(c, hw)
		return
	}

	screen.Main(func(s *screen.Screen, input <-chan interface{}, exitChan <-chan struct{}) {
		gb := NewGameBoy(c, s.GetOutputChannel(), hw, exitChan)
		go func() {
			for {
				select {
				case _, _ = <-exitChan:
					c.Shutdown()
					return
				case ev := <-input:
					switch e := ev.(type) {
					case screen.KeyEvent:
						if e.Key == sdl.K_d && e.Pressed {
							gb.PPU.PrintPalettes()
						}

						gb.Input.HandleKeyEvent(e.Pressed, e.Key)
					}
				}
			}
		}()

		noBootRom := *noboot ||
			(hw == consts.GBC && len(mmu.GBC_BOOTROM) == 0) ||
			(hw == consts.DMG && len(mmu.BOOTROM) == 0)

		gb.Init(noBootRom)
		gb.CPU.Dump = *dump
		gb.Run()
	})
}
