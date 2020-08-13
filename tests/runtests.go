package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

var (
	testMask = flag.String("mask", "", "Specify a mask of tests that should run")
)

func build(dir string) {
	cmd := exec.Command("go", "build")
	cmd.Dir = dir
	fmt.Println("Building...")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

type Hardware byte

const (
	None Hardware = iota
	Any
	DMG
	GBC
)

type TestDef struct {
	Mode Hardware
	Path []string
}

type TestDefs []TestDef

func (t TestDefs) Len() int {
	return len(t)
}

func (t TestDefs) Less(i, j int) bool {
	if t[i].Mode < t[j].Mode {
		return true
	}
	if t[i].Mode == t[j].Mode {
		return strings.Compare(filepath.Join(t[i].Path...), filepath.Join(t[j].Path...)) < 0
	}
	return false
}

func (t TestDefs) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func Test(mode Hardware, path ...string) TestDef {
	return TestDef{
		mode,
		path,
	}
}

func runMooneye(emuPath string, rom TestDef) {
	modeFlag := "-dmg"
	if rom.Mode == GBC {
		modeFlag = "-color"
	}

	testPath := filepath.Join(rom.Path...)
	if testMask != nil && (*testMask != "") {
		if matched, _ := filepath.Match(*testMask, testPath); !matched {
			return
		}
	}
	romFile := filepath.Join(pwd, "mooneye", testPath) + ".gb"
	romName := ""
	for _, p := range rom.Path {
		if romName == "" {
			romName = p
		} else {
			romName = romName + " -> " + p
		}
	}

	fmt.Printf("%-54s", "| `"+romName+"`")
	fmt.Print(" |")
	cmd := exec.Command(emuPath, "-mooneye", modeFlag, romFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 0 {
				fmt.Print("\033[0;31m ❌    \033[0;37m")
			}
		} else {
			panic(err)
		}
	} else {
		fmt.Print("\033[0;32m ✓     \033[0;37m")
	}
	fmt.Println(" |")
}

var mooneyeTests = TestDefs{
	Test(Any, "acceptance", "add_sp_e_timing"),
	Test(Any, "acceptance", "bits", "mem_oam"),
	Test(Any, "acceptance", "bits", "reg_f"),
	Test(DMG, "acceptance", "bits", "unused_hwio-GS"),
	Test(None, "acceptance", "boot_div-dmg0"),
	Test(DMG, "acceptance", "boot_div-dmgABCmgb"),
	Test(None, "acceptance", "boot_div-S"),
	Test(None, "acceptance", "boot_div2-S"),
	Test(None, "acceptance", "boot_hwio-dmg0"),
	Test(DMG, "acceptance", "boot_hwio-dmgABCmgb"),
	Test(None, "acceptance", "boot_hwio-S"),
	Test(None, "acceptance", "boot_regs-dmg0"),
	Test(DMG, "acceptance", "boot_regs-dmgABC"),
	Test(None, "acceptance", "boot_regs-mgb"),
	Test(None, "acceptance", "boot_regs-sgb"),
	Test(None, "acceptance", "boot_regs-sgb2"),
	Test(Any, "acceptance", "call_cc_timing"),
	Test(Any, "acceptance", "call_cc_timing2"),
	Test(Any, "acceptance", "call_timing"),
	Test(Any, "acceptance", "call_timing2"),
	Test(DMG, "acceptance", "di_timing-GS"),
	Test(Any, "acceptance", "div_timing"),
	Test(Any, "acceptance", "ei_sequence"),
	Test(Any, "acceptance", "ei_timing"),
	Test(Any, "acceptance", "halt_ime0_ei"),
	Test(Any, "acceptance", "halt_ime0_nointr_timing"),
	Test(Any, "acceptance", "halt_ime1_timing"),
	Test(DMG, "acceptance", "halt_ime1_timing2-GS"),
	Test(Any, "acceptance", "if_ie_registers"),
	Test(Any, "acceptance", "instr", "daa"),
	Test(Any, "acceptance", "interrupts", "ie_push"),
	Test(Any, "acceptance", "intr_timing"),
	Test(Any, "acceptance", "jp_cc_timing"),
	Test(Any, "acceptance", "jp_timing"),
	Test(Any, "acceptance", "ld_hl_sp_e_timing"),
	Test(Any, "acceptance", "oam_dma", "basic"),
	Test(Any, "acceptance", "oam_dma", "reg_read"),
	Test(DMG, "acceptance", "oam_dma", "sources-GS"),
	Test(Any, "acceptance", "oam_dma_restart"),
	Test(Any, "acceptance", "oam_dma_start"),
	Test(Any, "acceptance", "oam_dma_timing"),
	Test(Any, "acceptance", "pop_timing"),
	Test(DMG, "acceptance", "ppu", "hblank_ly_scx_timing-GS"),
	Test(DMG, "acceptance", "ppu", "intr_1_2_timing-GS"),
	Test(Any, "acceptance", "ppu", "intr_2_0_timing"),
	Test(Any, "acceptance", "ppu", "intr_2_mode0_timing"),
	Test(Any, "acceptance", "ppu", "intr_2_mode0_timing_sprites"),
	Test(Any, "acceptance", "ppu", "intr_2_mode3_timing"),
	Test(Any, "acceptance", "ppu", "intr_2_oam_ok_timing"),
	Test(DMG, "acceptance", "ppu", "lcdon_timing-GS"),
	Test(DMG, "acceptance", "ppu", "lcdon_write_timing-GS"),
	Test(Any, "acceptance", "ppu", "stat_irq_blocking"),
	Test(Any, "acceptance", "ppu", "stat_lyc_onoff"),
	Test(DMG, "acceptance", "ppu", "vblank_stat_intr-GS"),
	Test(Any, "acceptance", "push_timing"),
	Test(Any, "acceptance", "rapid_di_ei"),
	Test(Any, "acceptance", "ret_cc_timing"),
	Test(Any, "acceptance", "ret_timing"),
	Test(Any, "acceptance", "reti_intr_timing"),
	Test(Any, "acceptance", "reti_timing"),
	Test(Any, "acceptance", "rst_timing"),
	Test(DMG, "acceptance", "serial", "boot_sclk_align-dmgABCmgb"),
	Test(Any, "acceptance", "timer", "div_write"),
	Test(Any, "acceptance", "timer", "rapid_toggle"),
	Test(Any, "acceptance", "timer", "tim00"),
	Test(Any, "acceptance", "timer", "tim00_div_trigger"),
	Test(Any, "acceptance", "timer", "tim01"),
	Test(Any, "acceptance", "timer", "tim01_div_trigger"),
	Test(Any, "acceptance", "timer", "tim10"),
	Test(Any, "acceptance", "timer", "tim10_div_trigger"),
	Test(Any, "acceptance", "timer", "tim11"),
	Test(Any, "acceptance", "timer", "tim11_div_trigger"),
	Test(Any, "acceptance", "timer", "tima_reload"),
	Test(Any, "acceptance", "timer", "tima_write_reloading"),
	Test(Any, "acceptance", "timer", "tma_write_reloading"),
	Test(Any, "emulator-only", "mbc1", "bits_bank1"),
	Test(Any, "emulator-only", "mbc1", "bits_bank2"),
	Test(Any, "emulator-only", "mbc1", "bits_mode"),
	Test(Any, "emulator-only", "mbc1", "bits_ramg"),
	Test(Any, "emulator-only", "mbc1", "multicart_rom_8Mb"),
	Test(Any, "emulator-only", "mbc1", "ram_256kb"),
	Test(Any, "emulator-only", "mbc1", "ram_64kb"),
	Test(Any, "emulator-only", "mbc1", "rom_16Mb"),
	Test(Any, "emulator-only", "mbc1", "rom_1Mb"),
	Test(Any, "emulator-only", "mbc1", "rom_2Mb"),
	Test(Any, "emulator-only", "mbc1", "rom_4Mb"),
	Test(Any, "emulator-only", "mbc1", "rom_512kb"),
	Test(Any, "emulator-only", "mbc1", "rom_8Mb"),
	Test(Any, "emulator-only", "mbc2", "bits_ramg"),
	Test(Any, "emulator-only", "mbc2", "bits_romb"),
	Test(Any, "emulator-only", "mbc2", "bits_unused"),
	Test(Any, "emulator-only", "mbc2", "ram"),
	Test(Any, "emulator-only", "mbc2", "rom_1Mb"),
	Test(Any, "emulator-only", "mbc2", "rom_2Mb"),
	Test(Any, "emulator-only", "mbc2", "rom_512kb"),
	Test(Any, "emulator-only", "mbc5", "rom_16Mb"),
	Test(Any, "emulator-only", "mbc5", "rom_1Mb"),
	Test(Any, "emulator-only", "mbc5", "rom_2Mb"),
	Test(Any, "emulator-only", "mbc5", "rom_32Mb"),
	Test(Any, "emulator-only", "mbc5", "rom_4Mb"),
	Test(Any, "emulator-only", "mbc5", "rom_512kb"),
	Test(Any, "emulator-only", "mbc5", "rom_64Mb"),
	Test(Any, "emulator-only", "mbc5", "rom_8Mb"),
	Test(GBC, "misc", "bits", "unused_hwio-C"),
	Test(None, "misc", "boot_div-A"),
	Test(None, "misc", "boot_div-cgb0"),
	Test(GBC, "misc", "boot_div-cgbABCDE"),
	Test(GBC, "misc", "boot_hwio-C"),
	Test(None, "misc", "boot_regs-A"),
	Test(GBC, "misc", "boot_regs-cgb"),
	Test(GBC, "misc", "ppu", "vblank_stat_intr-C"),
}

var pwd string

func main() {
	flag.Parse()
	sort.Sort(mooneyeTests)
	var err error
	pwd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	build(filepath.Dir(pwd))

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	emu := filepath.Join(filepath.Dir(pwd), "goboy2"+ext)
	lastMode := None
	for _, met := range mooneyeTests {
		if met.Mode == None {
			continue
		}

		if met.Mode != lastMode {
			lastMode = met.Mode
			fmt.Println()
			fmt.Println()
			fmt.Print("\033[1;34m#### ")
			switch met.Mode {
			case Any:
				fmt.Print("General")
			case DMG:
				fmt.Print("DMG")
			case GBC:
				fmt.Print("GBC")
			}
			fmt.Println("\033[0m")
			fmt.Println()

			fmt.Println("| Test                                                 | Result |")
			fmt.Println("| ---------------------------------------------------- | ------ |")
		}
		runMooneye(emu, met)
	}
}
