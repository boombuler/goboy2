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
			romName = romName + "->" + p
		}
	}

	fmt.Printf("%-40s", romName+":")
	cmd := exec.Command(emuPath, "-mooneye", romFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

var mooneyeTests = [][]string{
	[]string{"bits", "mem_oam"},
	[]string{"bits", "reg_f"},
	[]string{"bits", "unused_hwio-GS"},
	[]string{"interrupts", "ie_push"},
	[]string{"oam_dma", "basic"},
	[]string{"oam_dma", "reg_read"},
	[]string{"ppu", "hblank_ly_scx_timing-GS"},
	[]string{"ppu", "intr_1_2_timing-GS"},
	[]string{"ppu", "intr_2_0_timing"},
	[]string{"ppu", "intr_2_mode0_timing"},
	[]string{"ppu", "intr_2_mode0_timing_sprites"},
	[]string{"ppu", "intr_2_mode3_timing"},
	[]string{"ppu", "intr_2_oam_ok_timing"},
	[]string{"ppu", "lcdon_write_timing-GS"},
	[]string{"ppu", "stat_irq_blocking"},
	[]string{"ppu", "stat_lyc_onoff"},
	[]string{"ppu", "vblank_stat_intr-GS"},
	[]string{"timer", "div_write"},
	[]string{"timer", "rapid_toggle"},
	[]string{"timer", "tim00"},
	[]string{"timer", "tim00_div_trigger"},
	[]string{"timer", "tim01"},
	[]string{"timer", "tim01_div_trigger"},
	[]string{"timer", "tim10"},
	[]string{"timer", "tim10_div_trigger"},
	[]string{"timer", "tim11"},
	[]string{"timer", "tim11_div_trigger"},
	[]string{"timer", "tima_reload"},
	[]string{"timer", "tima_write_reloading"},
	[]string{"timer", "tma_write_reloading"},
	[]string{"add_sp_e_timing"},
	[]string{"boot_hwio-dmg0"},
	[]string{"boot_regs-dmg0"},
	[]string{"call_cc_timing"},
	[]string{"call_cc_timing2"},
	[]string{"call_timing"},
	[]string{"call_timing2"},
	[]string{"di_timing-GS"},
	[]string{"div_timing"},
	[]string{"ei_sequence"},
	[]string{"ei_timing"},
	[]string{"halt_ime0_ei"},
	[]string{"halt_ime0_nointr_timing"},
	[]string{"halt_ime1_timing"},
	[]string{"halt_ime1_timing2-GS"},
	[]string{"if_ie_registers"},
	[]string{"intr_timing"},
	[]string{"jp_cc_timing"},
	[]string{"jp_timing"},
	[]string{"ld_hl_sp_e_timing"},
	[]string{"oam_dma_restart"},
	[]string{"oam_dma_start"},
	[]string{"oam_dma_timing"},
	[]string{"pop_timing"},
	[]string{"push_timing"},
	[]string{"rapid_di_ei"},
	[]string{"ret_cc_timing"},
	[]string{"ret_timing"},
	[]string{"reti_intr_timing"},
	[]string{"reti_timing"},
	[]string{"rst_timing"},
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
