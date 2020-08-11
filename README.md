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