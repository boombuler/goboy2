package consts

const (
	// Freq is the number of T-Cycles per second
	Freq = 4194304
	// TicksPerSecond is the number of M-Cycles per second
	TicksPerSecond = Freq / 4

	// DisplayWidth is the width of the output pictures
	DisplayWidth int = 160

	// DisplayHeight is the height of the output pictures
	DisplayHeight int = 144
)

const (
	// AddrInput is the address of the Input (Gamepad) HardwareRegister
	AddrInput uint16 = 0xFF00
	// AddrDivider is the address of the timer DIV register
	AddrDivider uint16 = 0xFF04
	// AddrTIMA is the address of the timer TIMA register
	AddrTIMA uint16 = 0xFF05
	// AddrModulo is the address of the timer TMA register
	AddrModulo uint16 = 0xFF06
	// AddrCtrl is the address of the timer TAC register
	AddrCtrl uint16 = 0xFF07
	// AddrIRQFlags is the address of the IF HardwareRegister
	AddrIRQFlags uint16 = 0xFF0F

	AddrLCDC    uint16 = 0xFF40
	AddrSTAT    uint16 = 0xFF41
	AddrSCROLLY uint16 = 0xFF42
	AddrSCROLLX uint16 = 0xFF43
	AddrLY      uint16 = 0xFF44
	AddrLYC     uint16 = 0xFF45
	// AddrDMATransfer is the address of the OAM-DMA Transfer Control HardwareRegister
	AddrDMATransfer    uint16 = 0xFF46
	AddrBGP            uint16 = 0xFF47
	AddrOBJECTPALETTE0 uint16 = 0xFF48
	AddrOBJECTPALETTE1 uint16 = 0xFF49
	AddrWY             uint16 = 0xFF4A
	AddrWX             uint16 = 0xFF4B
	AddrLCDMODE        uint16 = 0xFF4C

	AddrVBK   uint16 = 0xFF4F
	AddrHDMA1 uint16 = 0xFF51
	AddrHDMA2 uint16 = 0xFF52
	AddrHDMA3 uint16 = 0xFF53
	AddrHDMA4 uint16 = 0xFF54
	AddrHDMA5 uint16 = 0xFF55

	AddrBGPI uint16 = 0xFF68
	AddrBGPD uint16 = 0xFF69
	AddrOBPI uint16 = 0xFF6A
	AddrOBPD uint16 = 0xFF6B

	// AddrKEY1 is the GBC Speed-Mode Register
	AddrKEY1 uint16 = 0xFF4D
	// AddrRP is the GBC  Infrared Communications Port
	AddrRP uint16 = 0xFF56
	// AddrBootmodeFlag is the address of the BootMode HardwareRegister
	AddrBootmodeFlag uint16 = 0xFF50
	// AddrSVBK is the working ram bank seletor for GBC
	AddrSVBK uint16 = 0xFF70
	// AddrIRQEnabled is the address of the IE HardwareRegister
	AddrIRQEnabled uint16 = 0xFFFF
)
