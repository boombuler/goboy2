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

	// AddrBootmodeFlag is the address of the BootMode HardwareRegister
	AddrBootmodeFlag uint16 = 0xFF50
	// AddrIRQEnabled is the address of the IE HardwareRegister
	AddrIRQEnabled uint16 = 0xFFFF
)