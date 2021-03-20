package i2c

import (
	"errors"

	"gobot.io/x/gobot"
)

// PCA953xAddress is set to variant PCA9533/2
const PCA953xAddress = 0x63

// there are 6 registers
const (
	PCA953xRegInp  = 0x00 // input register
	PCA953xRegPsc0 = 0x01 // r,   frequency prescaler 0
	PCA953xRegPwm0 = 0x02 // r/w, PWM register 0
	PCA953xRegPsc1 = 0x03 // r/w, frequency prescaler 1
	PCA953xRegPwm1 = 0x04 // r/w, PWM register 1
	PCA953xRegLs0  = 0x05 // r/w, LED selector 0
	PCA953xRegLs1  = 0x06 // r/w, LED selector 1 (only in PCA9531, PCA9532)
)

const PCA953xAiMask = 0x10 // autoincrement bit

var ErrToMuchBytes = errors.New("To much bytes read")

// PCA953x is a Gobot Driver for LED Dimmer PCA9530 (2-bit), PCA9533 (4-bit), PCA9531 (8-bit), PCA9532 (16-bit)
//
// Address range:
// * PCA9530   0x60-0x61 (96-97 dec)
// * PCA9531   0x60-0x67 (96-103 dec)
// * PCA9532   0x60-0x67 (96-103 dec)
// * PCA9533/1 0x62      (98 dec)
// * PCA9533/2 0x63      (99 dec)
//
// each new command must start by setting the register and the AI flag
// 0 0 0 AI | 0 R2 R1 R0
// AI=1 means autoincrementing R0-R2, which enable reading/writing all registers sequencially
// when AI=1 and reading, then R!=0
// this means: do not start with reading input register, writing input register is recognized but has no effect
// => when AI=1 in general start with R>0
//
type PCA953xDriver struct {
	name       string
	connector  Connector
	connection Connection
	Config
	gobot.Commander
}

// NewPCA953xDriver creates a new driver with specified i2c interface
// Params:
//		conn Connector - the Adaptor to use with this Driver
//
// Optional params:
//		i2c.WithBus(int):	bus to use with this driver
//		i2c.WithAddress(int):	address to use with this driver
//
func NewPCA953xDriver(a Connector, options ...func(Config)) *PCA953xDriver {
	p := &PCA953xDriver{
		name:      gobot.DefaultName("PCA953x"),
		connector: a,
		Config:    NewConfig(),
		Commander: gobot.NewCommander(),
	}

	for _, option := range options {
		option(p)
	}

	// TODO: API commands
	return p
}

// Name returns the Name for the Driver
func (p *PCA953xDriver) Name() string { return p.name }

// SetName sets the Name for the Driver
func (p *PCA953xDriver) SetName(n string) { p.name = n }

// Connection returns the connection for the Driver
func (p *PCA953xDriver) Connection() gobot.Connection { return p.connector.(gobot.Connection) }

// Start initializes the PCA953x
func (p *PCA953xDriver) Start() (err error) {
	bus := p.GetBusOrDefault(p.connector.GetDefaultBus())
	address := p.GetAddressOrDefault(PCA953xAddress)
	p.connection, err = p.connector.GetConnection(address, bus)
	return err
}

// Halt stops the device
func (p *PCA953xDriver) Halt() (err error) { return }

// InputRegister get the content of the input register
func (p *PCA953xDriver) InputRegister() (uint8, error) {
	return p.readRegister(PCA953xRegInp)
}

// Psc0Register get the content of the frequency prescaler 0 (PSC0)
func (p *PCA953xDriver) Psc0Register() (uint8, error) {
	return p.readRegister(PCA953xRegPsc0)
}

// Psc1Register get the content of the frequency prescaler 1 (PSC1)
func (p *PCA953xDriver) Psc1Register() (uint8, error) {
	return p.readRegister(PCA953xRegPsc1)
}

// Pwm0Register get the content of the pulse wide modulation (PWM0)
func (p *PCA953xDriver) Pwm0Register() (uint8, error) {
	return p.readRegister(PCA953xRegPwm0)
}

// Pwm1Register get the content of the pulse wide modulation (PWM1)
func (p *PCA953xDriver) Pwm1Register() (uint8, error) {
	return p.readRegister(PCA953xRegPwm1)
}

// Ls0Register get the content of the LED selector 0 (LS0)
func (p *PCA953xDriver) Ls0Register() (uint8, error) {
	return p.readRegister(PCA953xRegLs0)
}

// Ls1Register get the content of the LED selector 1 (LS1)
func (p *PCA953xDriver) Ls1Register() (uint8, error) {
	return p.readRegister(PCA953xRegLs1)
}

// WritePsc0Register set the content of the frequency prescaler 0 (PSC0)
func (p *PCA953xDriver) WritePsc0Register(val uint8) error {
	return p.writeRegister(PCA953xRegPsc0, val)
}

// WritePsc1Register set the content of the frequency prescaler 1 (PSC1)
func (p *PCA953xDriver) WritePsc1Register(val uint8) error {
	return p.writeRegister(PCA953xRegPsc1, val)
}

// WritePwm0Register set the content of the pulse wide modulation (PWM0)
func (p *PCA953xDriver) WritePwm0Register(val uint8) error {
	return p.writeRegister(PCA953xRegPwm0, val)
}

// WritePwm1Register set the content of the pulse wide modulation (PWM1)
func (p *PCA953xDriver) WritePwm1Register(val uint8) error {
	return p.writeRegister(PCA953xRegPwm1, val)
}

// WriteLs0Register set the content of the LED selector 0 (LS0)
func (p *PCA953xDriver) WriteLs0Register(val uint8) error {
	return p.writeRegister(PCA953xRegLs0, val)
}

// WriteLs1Register set the content of the LED selector 1 (LS1)
func (p *PCA953xDriver) WriteLs1Register(val uint8) error {
	return p.writeRegister(PCA953xRegLs1, val)
}

// read the content of the given register
func (p *PCA953xDriver) readRegister(regAddress uint8) (uint8, error) {
	// ensure AI bit is not set
	regAddress = regAddress &^ PCA953xAiMask
	// write CTRL register
	err := p.write(uint8(regAddress))
	if err != nil {
		return 0, err
	}
	// read content of requested register
	return p.read()
}

// write the content of the given register
func (p *PCA953xDriver) writeRegister(regAddress uint8, val uint8) error {
	// ensure AI bit is not set
	regAddress = regAddress &^ PCA953xAiMask
	// write CTRL register
	err := p.write(uint8(regAddress))
	if err != nil {
		return err
	}
	// write content of requested register
	return p.write(val)
}

// write the value to the connection
func (p *PCA953xDriver) write(val uint8) error {
	_, err := p.connection.Write([]uint8{val})
	return err
}

// read the value from the connection
func (p *PCA953xDriver) read() (uint8, error) {
	buf := []byte{0}
	bytesRead, err := p.connection.Read(buf)
	if err != nil {
		return 0, err
	}
	if bytesRead < 1 {
		return 0, ErrNotEnoughBytes
	}
	if bytesRead > 1 {
		return 0, ErrToMuchBytes
	}
	return buf[0], nil
}
