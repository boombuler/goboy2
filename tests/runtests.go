package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func build(dir string) {
	cmd := exec.Command("go", "build")
	cmd.Dir = dir
	fmt.Println("Building...")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func runMooneye(emuPath string, rom []string) {
	romFile := filepath.Join(pwd, "mooneye", filepath.Join(rom...)) + ".gb"
	romName := ""
	for _, p := range rom {
		if romName == "" {
			romName = p
		} else {
			romName = romName + " -> " + p
		}
	}

	fmt.Printf("%-50s", romName+":")
	cmd := exec.Command(emuPath, "-mooneye", romFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

var mooneyeTests = [][]string{
	[]string{"acceptance", "add_sp_e_timing"},
	[]string{"acceptance", "bits", "mem_oam"},
	[]string{"acceptance", "bits", "reg_f"},
	[]string{"acceptance", "bits", "unused_hwio-GS"},
	[]string{"acceptance", "boot_div-dmg0"},
	[]string{"acceptance", "boot_div-dmgABCmgb"},
	[]string{"acceptance", "boot_div-S"},
	[]string{"acceptance", "boot_div2-S"},
	[]string{"acceptance", "boot_hwio-dmg0"},
	[]string{"acceptance", "boot_hwio-dmgABCmgb"},
	[]string{"acceptance", "boot_hwio-S"},
	[]string{"acceptance", "boot_regs-dmg0"},
	[]string{"acceptance", "boot_regs-dmgABC"},
	[]string{"acceptance", "boot_regs-mgb"},
	[]string{"acceptance", "boot_regs-sgb"},
	[]string{"acceptance", "boot_regs-sgb2"},
	[]string{"acceptance", "call_cc_timing"},
	[]string{"acceptance", "call_cc_timing2"},
	[]string{"acceptance", "call_timing"},
	[]string{"acceptance", "call_timing2"},
	[]string{"acceptance", "di_timing-GS"},
	[]string{"acceptance", "div_timing"},
	[]string{"acceptance", "ei_sequence"},
	[]string{"acceptance", "ei_timing"},
	[]string{"acceptance", "halt_ime0_ei"},
	[]string{"acceptance", "halt_ime0_nointr_timing"},
	[]string{"acceptance", "halt_ime1_timing"},
	[]string{"acceptance", "halt_ime1_timing2-GS"},
	[]string{"acceptance", "if_ie_registers"},
	[]string{"acceptance", "instr", "daa"},
	[]string{"acceptance", "interrupts", "ie_push"},
	[]string{"acceptance", "intr_timing"},
	[]string{"acceptance", "jp_cc_timing"},
	[]string{"acceptance", "jp_timing"},
	[]string{"acceptance", "ld_hl_sp_e_timing"},
	[]string{"acceptance", "oam_dma", "basic"},
	[]string{"acceptance", "oam_dma", "reg_read"},
	[]string{"acceptance", "oam_dma", "sources-GS"},
	[]string{"acceptance", "oam_dma_restart"},
	[]string{"acceptance", "oam_dma_start"},
	[]string{"acceptance", "oam_dma_timing"},
	[]string{"acceptance", "pop_timing"},
	[]string{"acceptance", "ppu", "hblank_ly_scx_timing-GS"},
	[]string{"acceptance", "ppu", "intr_1_2_timing-GS"},
	[]string{"acceptance", "ppu", "intr_2_0_timing"},
	[]string{"acceptance", "ppu", "intr_2_mode0_timing"},
	[]string{"acceptance", "ppu", "intr_2_mode0_timing_sprites"},
	[]string{"acceptance", "ppu", "intr_2_mode3_timing"},
	[]string{"acceptance", "ppu", "intr_2_oam_ok_timing"},
	[]string{"acceptance", "ppu", "lcdon_timing-GS"},
	[]string{"acceptance", "ppu", "lcdon_write_timing-GS"},
	[]string{"acceptance", "ppu", "stat_irq_blocking"},
	[]string{"acceptance", "ppu", "stat_lyc_onoff"},
	[]string{"acceptance", "ppu", "vblank_stat_intr-GS"},
	[]string{"acceptance", "push_timing"},
	[]string{"acceptance", "rapid_di_ei"},
	[]string{"acceptance", "ret_cc_timing"},
	[]string{"acceptance", "ret_timing"},
	[]string{"acceptance", "reti_intr_timing"},
	[]string{"acceptance", "reti_timing"},
	[]string{"acceptance", "rst_timing"},
	[]string{"acceptance", "serial", "boot_sclk_align-dmgABCmgb"},
	[]string{"acceptance", "timer", "div_write"},
	[]string{"acceptance", "timer", "rapid_toggle"},
	[]string{"acceptance", "timer", "tim00"},
	[]string{"acceptance", "timer", "tim00_div_trigger"},
	[]string{"acceptance", "timer", "tim01"},
	[]string{"acceptance", "timer", "tim01_div_trigger"},
	[]string{"acceptance", "timer", "tim10"},
	[]string{"acceptance", "timer", "tim10_div_trigger"},
	[]string{"acceptance", "timer", "tim11"},
	[]string{"acceptance", "timer", "tim11_div_trigger"},
	[]string{"acceptance", "timer", "tima_reload"},
	[]string{"acceptance", "timer", "tima_write_reloading"},
	[]string{"acceptance", "timer", "tma_write_reloading"},
	[]string{"emulator-only", "mbc1", "bits_bank1"},
	[]string{"emulator-only", "mbc1", "bits_bank2"},
	[]string{"emulator-only", "mbc1", "bits_mode"},
	[]string{"emulator-only", "mbc1", "bits_ramg"},
	[]string{"emulator-only", "mbc1", "multicart_rom_8Mb"},
	[]string{"emulator-only", "mbc1", "ram_256kb"},
	[]string{"emulator-only", "mbc1", "ram_64kb"},
	[]string{"emulator-only", "mbc1", "rom_16Mb"},
	[]string{"emulator-only", "mbc1", "rom_1Mb"},
	[]string{"emulator-only", "mbc1", "rom_2Mb"},
	[]string{"emulator-only", "mbc1", "rom_4Mb"},
	[]string{"emulator-only", "mbc1", "rom_512kb"},
	[]string{"emulator-only", "mbc1", "rom_8Mb"},
	[]string{"emulator-only", "mbc2", "bits_ramg"},
	[]string{"emulator-only", "mbc2", "bits_romb"},
	[]string{"emulator-only", "mbc2", "bits_unused"},
	[]string{"emulator-only", "mbc2", "ram"},
	[]string{"emulator-only", "mbc2", "rom_1Mb"},
	[]string{"emulator-only", "mbc2", "rom_2Mb"},
	[]string{"emulator-only", "mbc2", "rom_512kb"},
	[]string{"emulator-only", "mbc5", "rom_16Mb"},
	[]string{"emulator-only", "mbc5", "rom_1Mb"},
	[]string{"emulator-only", "mbc5", "rom_2Mb"},
	[]string{"emulator-only", "mbc5", "rom_32Mb"},
	[]string{"emulator-only", "mbc5", "rom_4Mb"},
	[]string{"emulator-only", "mbc5", "rom_512kb"},
	[]string{"emulator-only", "mbc5", "rom_64Mb"},
	[]string{"emulator-only", "mbc5", "rom_8Mb"},
	[]string{"misc", "bits", "unused_hwio-C"},
	[]string{"misc", "boot_div-A"},
	[]string{"misc", "boot_div-cgb0"},
	[]string{"misc", "boot_div-cgbABCDE"},
	[]string{"misc", "boot_hwio-C"},
	[]string{"misc", "boot_regs-A"},
	[]string{"misc", "boot_regs-cgb"},
	[]string{"misc", "ppu", "vblank_stat_intr-C"},
}

var pwd string

func main() {
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
	for _, met := range mooneyeTests {
		runMooneye(emu, met)
	}
}
