package i2c

import (
	"errors"

	"gobot.io/x/gobot"
)

// PCA953xAddress is set to variant PCA9533/2
const PCA953xAddress = 0x63

// there are 6 registers
const (
	pca953xRegInp  = 0x00 // input register
	pca953xRegPsc0 = 0x01 // r,   frequency prescaler 0
	pca953xRegPwm0 = 0x02 // r/w, PWM register 0
	pca953xRegPsc1 = 0x03 // r/w, frequency prescaler 1
	pca953xRegPwm1 = 0x04 // r/w, PWM register 1
	pca953xRegLs0  = 0x05 // r/w, LED selector 0
	pca953xRegLs1  = 0x06 // r/w, LED selector 1 (only in PCA9531, PCA9532)
)

// autoincrement bit
const pca953xAiMask = 0x10

// PCA953xGPIOMode is used to set the mode while write GPIO
type PCA953xGPIOMode uint8

const (
	// PCA953xModeHigh set the GPIO to high (LED off)
	PCA953xModeHigh PCA953xGPIOMode = 0x00
	// PCA953xModeLow set the GPIO to low (LED on)
	PCA953xModeLow = 0x01
	// PCA953xModePwm0 set the GPIO to PWM (PWM0 & PSC0)
	PCA953xModePwm0 = 0x02
	// PCA953xModePwm0 set the GPIO to PWM (PWM1 & PSC1)
	PCA953xModePwm1 = 0x03
)

var ErrToMuchBytes = errors.New("To much bytes read")

// PCA953x is a Gobot Driver for LED Dimmer PCA9530 (2-bit), PCA9533 (4-bit), PCA9531 (8-bit), PCA9532 (16-bit)
// Although this is designed for LED's it can be used as a GPIO (read, write, pwm).
// The names of the public functions reflect this.
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

// WriteGPIO writes a value to a gpio output (index 0-7)
func (p *PCA953xDriver) WriteGPIO(idx uint8, mode PCA953xGPIOMode) (err error) {
	// prepare
	var regLs uint8 = pca953xRegLs0
	if idx > 3 {
		regLs = pca953xRegLs1
		idx = idx - 4
	}
	regLsShift := idx * 2
	// read old value
	regLsVal, err := p.readRegister(regLs)
	// reset 2 bits at LED postion
	regLsVal &= ^uint8(0x03 << regLsShift)
	// set 2 bits according to mode at LED position
	regLsVal |= uint8(mode) << regLsShift
	// write new value
	return p.writeRegister(regLs, regLsVal)
}

// ReadGPIO reads a gpio input (index 0-7) to a value
func (p *PCA953xDriver) ReadGPIO(idx uint8) (uint8, error) {
	// read input register
	val, err := p.readRegister(pca953xRegInp)
	// create return bit
	if err != nil {
		return val, err
	}
	val = 1 << uint8(idx) & val
	if val > 1 {
		val = 1
	}
	return val, nil
}

// WritePsc0Register set the content of the frequency prescaler 0 (PSC0)
func (p *PCA953xDriver) WritePsc0Register(val uint8) error {
	return p.writeRegister(pca953xRegPsc0, val)
}

// WritePsc1Register set the content of the frequency prescaler 1 (PSC1)
func (p *PCA953xDriver) WritePsc1Register(val uint8) error {
	return p.writeRegister(pca953xRegPsc1, val)
}

// WritePwm0Register set the content of the pulse wide modulation (PWM0)
func (p *PCA953xDriver) WritePwm0Register(val uint8) error {
	return p.writeRegister(pca953xRegPwm0, val)
}

// WritePwm1Register set the content of the pulse wide modulation (PWM1)
func (p *PCA953xDriver) WritePwm1Register(val uint8) error {
	return p.writeRegister(pca953xRegPwm1, val)
}

// Psc0Register get the content of the frequency prescaler 0 (PSC0)
func (p *PCA953xDriver) Psc0Register() (uint8, error) {
	return p.readRegister(pca953xRegPsc0)
}

// Psc1Register get the content of the frequency prescaler 1 (PSC1)
func (p *PCA953xDriver) Psc1Register() (uint8, error) {
	return p.readRegister(pca953xRegPsc1)
}

// Pwm0Register get the content of the pulse wide modulation (PWM0)
func (p *PCA953xDriver) Pwm0Register() (uint8, error) {
	return p.readRegister(pca953xRegPwm0)
}

// Pwm1Register get the content of the pulse wide modulation (PWM1)
func (p *PCA953xDriver) Pwm1Register() (uint8, error) {
	return p.readRegister(pca953xRegPwm1)
}

// write the content of the given register
func (p *PCA953xDriver) writeRegister(regAddress uint8, val uint8) error {
	// ensure AI bit is not set
	regAddress = regAddress &^ pca953xAiMask
	// write CTRL register
	err := p.write(uint8(regAddress))
	if err != nil {
		return err
	}
	// write content of requested register
	return p.write(val)
}

// read the content of the given register
func (p *PCA953xDriver) readRegister(regAddress uint8) (uint8, error) {
	// ensure AI bit is not set
	regAddress = regAddress &^ pca953xAiMask
	// write CTRL register
	err := p.write(uint8(regAddress))
	if err != nil {
		return 0, err
	}
	// read content of requested register
	return p.read()
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
