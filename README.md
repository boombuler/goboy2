# GOBOY2

## Introduction

As the name implies this is my second try to develop a gameboy emulator in go. 
The first one was capable of playing Tetris and even some other games, but I wanted to improve
the accuracy of emulation. To do so I've decided to rewrite the CPU to use some kind of *micro-opcodes*
which really helped a lot.

## Boot-Roms

Due to the fact, that the data of the rom chips of the original gameboy are IP of Nintendo. This data
is stripped from the code. It could be added by adding the binary data to `mmu/bootom.go`. 
The other option is to start the emulator using the `-noboot` option.

## Deployment

You might need to install the SDL2 libs for your system. 

## Input

The default key-map is as follows:

```
	Up, Down, 
	Left, Right --> Arrow Keys

	Start       --> Return 
	Select      --> Backspace
	
	A           --> X-Key
	B           --> Y-Key
```

## Tests

### Blargg

| Test             | Result             |
| ---------------- | ------------------ |
| `cgb_sound`      | :x:                |
| `cpu_instrs`     | :heavy_check_mark: |
| `dmg_sound`      | :x:                |
| `instr_timing`   | :heavy_check_mark: |
| `interrupt_time` | :x:                |
| `mem_timing`     | :heavy_check_mark: |
| `mem_timing-2`   | :heavy_check_mark: |
| `oam_bug`        | :x:                |
| `halt_bug`       | :heavy_check_mark: |

### Mooneye

With the [tests from the 18. Apr. 2020](https://github.com/Gekkio/mooneye-gb/commit/6b9488fa3e7da033a3c33c55ac94476c0e8368b0) the emulator currently gets the following results:


| Test                                                | Result             | 
| --------------------------------------------------- | ------------------ | 
| `acceptance -> add_sp_e_timing`                     | :heavy_check_mark: |
| `acceptance -> bits -> mem_oam`                     | :heavy_check_mark: |
| `acceptance -> bits -> reg_f`                       | :heavy_check_mark: |
| `acceptance -> bits -> unused_hwio-GS`              | :heavy_check_mark: |
| `acceptance -> boot_div-dmg0`                       | :x:                |
| `acceptance -> boot_div-dmgABCmgb`                  | :heavy_check_mark: |
| `acceptance -> boot_div-S`                          | :x:                |
| `acceptance -> boot_div2-S`                         | :x:                |
| `acceptance -> boot_hwio-dmg0`                      | :x:                |
| `acceptance -> boot_hwio-dmgABCmgb`                 | :heavy_check_mark: |
| `acceptance -> boot_hwio-S`                         | :x:                |
| `acceptance -> boot_regs-dmg0`                      | :x:                |
| `acceptance -> boot_regs-dmgABC`                    | :heavy_check_mark: |
| `acceptance -> boot_regs-mgb`                       | :x:                |
| `acceptance -> boot_regs-sgb`                       | :x:                |
| `acceptance -> boot_regs-sgb2`                      | :x:                |
| `acceptance -> call_cc_timing`                      | :heavy_check_mark: |
| `acceptance -> call_cc_timing2`                     | :heavy_check_mark: |
| `acceptance -> call_timing`                         | :heavy_check_mark: |
| `acceptance -> call_timing2`                        | :heavy_check_mark: |
| `acceptance -> di_timing-GS`                        | :x:                |
| `acceptance -> div_timing`                          | :heavy_check_mark: |
| `acceptance -> ei_sequence`                         | :heavy_check_mark: |
| `acceptance -> ei_timing`                           | :heavy_check_mark: |
| `acceptance -> halt_ime0_ei`                        | :heavy_check_mark: |
| `acceptance -> halt_ime0_nointr_timing`             | :x:                |
| `acceptance -> halt_ime1_timing`                    | :heavy_check_mark: |
| `acceptance -> halt_ime1_timing2-GS`                | :x:                |
| `acceptance -> if_ie_registers`                     | :heavy_check_mark: |
| `acceptance -> instr -> daa`                        | :heavy_check_mark: |
| `acceptance -> interrupts -> ie_push`               | :heavy_check_mark: |
| `acceptance -> intr_timing`                         | :heavy_check_mark: |
| `acceptance -> jp_cc_timing`                        | :heavy_check_mark: |
| `acceptance -> jp_timing`                           | :heavy_check_mark: |
| `acceptance -> ld_hl_sp_e_timing`                   | :heavy_check_mark: |
| `acceptance -> oam_dma -> basic`                    | :heavy_check_mark: |
| `acceptance -> oam_dma -> reg_read`                 | :heavy_check_mark: |
| `acceptance -> oam_dma -> sources-GS`               | :x:                |
| `acceptance -> oam_dma_restart`                     | :heavy_check_mark: |
| `acceptance -> oam_dma_start`                       | :heavy_check_mark: |
| `acceptance -> oam_dma_timing`                      | :heavy_check_mark: |
| `acceptance -> pop_timing`                          | :heavy_check_mark: |
| `acceptance -> ppu -> hblank_ly_scx_timing-GS`      | :x:                |
| `acceptance -> ppu -> intr_1_2_timing-GS`           | :x:                |
| `acceptance -> ppu -> intr_2_0_timing`              | :x:                |
| `acceptance -> ppu -> intr_2_mode0_timing`          | :x:                |
| `acceptance -> ppu -> intr_2_mode0_timing_sprites`  | :x:                |
| `acceptance -> ppu -> intr_2_mode3_timing`          | :x:                |
| `acceptance -> ppu -> intr_2_oam_ok_timing`         | :x:                |
| `acceptance -> ppu -> lcdon_timing-GS`              | :x:                |
| `acceptance -> ppu -> lcdon_write_timing-GS`        | :x:                |
| `acceptance -> ppu -> stat_irq_blocking`            | :x:                |
| `acceptance -> ppu -> stat_lyc_onoff`               | :x:                |
| `acceptance -> ppu -> vblank_stat_intr-GS`          | :x:                |
| `acceptance -> push_timing`                         | :heavy_check_mark: |
| `acceptance -> rapid_di_ei`                         | :x:                |
| `acceptance -> ret_cc_timing`                       | :heavy_check_mark: |
| `acceptance -> ret_timing`                          | :heavy_check_mark: |
| `acceptance -> reti_intr_timing`                    | :heavy_check_mark: |
| `acceptance -> reti_timing`                         | :heavy_check_mark: |
| `acceptance -> rst_timing`                          | :heavy_check_mark: |
| `acceptance -> serial -> boot_sclk_align-dmgABCmgb` | :x:                |
| `acceptance -> timer -> div_write`                  | :heavy_check_mark: |
| `acceptance -> timer -> rapid_toggle`               | :heavy_check_mark: |
| `acceptance -> timer -> tim00`                      | :heavy_check_mark: |
| `acceptance -> timer -> tim00_div_trigger`          | :heavy_check_mark: |
| `acceptance -> timer -> tim01`                      | :heavy_check_mark: |
| `acceptance -> timer -> tim01_div_trigger`          | :heavy_check_mark: |
| `acceptance -> timer -> tim10`                      | :heavy_check_mark: |
| `acceptance -> timer -> tim10_div_trigger`          | :heavy_check_mark: |
| `acceptance -> timer -> tim11`                      | :heavy_check_mark: |
| `acceptance -> timer -> tim11_div_trigger`          | :heavy_check_mark: |
| `acceptance -> timer -> tima_reload`                | :heavy_check_mark: |
| `acceptance -> timer -> tima_write_reloading`       | :heavy_check_mark: |
| `acceptance -> timer -> tma_write_reloading`        | :heavy_check_mark: |
| `emulator-only -> mbc1 -> bits_bank1`               | :heavy_check_mark: |
| `emulator-only -> mbc1 -> bits_bank2`               | :heavy_check_mark: |
| `emulator-only -> mbc1 -> bits_mode`                | :heavy_check_mark: |
| `emulator-only -> mbc1 -> bits_ramg`                | :heavy_check_mark: |
| `emulator-only -> mbc1 -> multicart_rom_8Mb`        | :heavy_check_mark: |
| `emulator-only -> mbc1 -> ram_256kb`                | :heavy_check_mark: |
| `emulator-only -> mbc1 -> ram_64kb`                 | :heavy_check_mark: |
| `emulator-only -> mbc1 -> rom_16Mb`                 | :heavy_check_mark: |
| `emulator-only -> mbc1 -> rom_1Mb`                  | :heavy_check_mark: |
| `emulator-only -> mbc1 -> rom_2Mb`                  | :heavy_check_mark: |
| `emulator-only -> mbc1 -> rom_4Mb`                  | :heavy_check_mark: |
| `emulator-only -> mbc1 -> rom_512kb`                | :heavy_check_mark: |
| `emulator-only -> mbc1 -> rom_8Mb`                  | :heavy_check_mark: |
| `emulator-only -> mbc2 -> bits_ramg`                | :heavy_check_mark: |
| `emulator-only -> mbc2 -> bits_romb`                | :heavy_check_mark: |
| `emulator-only -> mbc2 -> bits_unused`              | :heavy_check_mark: |
| `emulator-only -> mbc2 -> ram`                      | :heavy_check_mark: |
| `emulator-only -> mbc2 -> rom_1Mb`                  | :heavy_check_mark: |
| `emulator-only -> mbc2 -> rom_2Mb`                  | :heavy_check_mark: |
| `emulator-only -> mbc2 -> rom_512kb`                | :heavy_check_mark: |
| `emulator-only -> mbc5 -> rom_16Mb`                 | :heavy_check_mark: |
| `emulator-only -> mbc5 -> rom_1Mb`                  | :heavy_check_mark: |
| `emulator-only -> mbc5 -> rom_2Mb`                  | :heavy_check_mark: |
| `emulator-only -> mbc5 -> rom_32Mb`                 | :heavy_check_mark: |
| `emulator-only -> mbc5 -> rom_4Mb`                  | :heavy_check_mark: |
| `emulator-only -> mbc5 -> rom_512kb`                | :heavy_check_mark: |
| `emulator-only -> mbc5 -> rom_64Mb`                 | :heavy_check_mark: |
| `emulator-only -> mbc5 -> rom_8Mb`                  | :heavy_check_mark: |
| `misc -> bits -> unused_hwio-C`                     | :x:                |
| `misc -> boot_div-A`                                | :x:                |
| `misc -> boot_div-cgb0`                             | :x:                |
| `misc -> boot_div-cgbABCDE`                         | :x:                |
| `misc -> boot_hwio-C`                               | :x:                |
| `misc -> boot_regs-A`                               | :x:                |
| `misc -> boot_regs-cgb`                             | :x:                |
| `misc -> ppu -> vblank_stat_intr-C`                 | :x:                |