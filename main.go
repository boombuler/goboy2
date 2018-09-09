package main

import (
	"flag"
	"goboy2/cartridge"
	"goboy2/screen"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
)

func showUsage() {
	log.Println("Usage:")
	log.Println(path.Base(os.Args[0]), "(romfile)")
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

var (
	dump       = flag.Bool("dump", false, "dump cpu state after every instruction")
	noboot     = flag.Bool("noboot", false, "skip boot sequence")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

func main() {
	flag.Parse()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
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

	screen.Main(func(s *screen.Screen, input <-chan interface{}, exitChan <-chan struct{}) {
		gb := NewGameBoy(c)
		go func() {
			for {
				select {
				case _, _ = <-exitChan:
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
		if err := gb.APU.Start(); err != nil {
			panic(err)
		}
		gb.CPU.Dump = *dump
		for {
			img := gb.StepFrame()
			select {
			case _, _ = <-exitChan:
				gb.APU.Stop()
			default:
				s.Show(img)
			}
		}
	})
}
