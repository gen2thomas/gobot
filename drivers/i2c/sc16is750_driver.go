package i2c

import (
	"gobot.io/x/gobot"
)

// Because read of registers and write GPIO's not working (at least with my device) I stop development at this point.
// Tested successfully:
// * Device was detected
// * Write of registers runs without an error
// * change of device address is recognized
//
// Tested without success at system level with my device (without gobot):
// * i2cset -y -r 6 0x4d 0x0B 0xFF
// * i2cset -y -r 6 0x4d 0x0B 0x00
//
// This normally should change the state of output and also the read back of register should match, but doesn't.
// Also set of direction register before set state don't change anything.
// So I suppose my device (breakout board similar to sparkfun 9981) is defective or I don't understand the usage information provided by NXP.
//

// sc16is750 has 2 address pins and supports 16 addresses from 0x48 to 0x57 by connecting pins to VDD, VSS, SCL, SDA
// A connection seems to be mandatory, because there are no internal resistors.
// Default address 0x4D is for both pins are grounded.
const sc16is750Address = 0x4d

// sc16is750Register is used to specify the register
type sc16is750Register uint8

// there are 15 registers
const (
	sc16is750RegRhrThrDll   sc16is750Register = iota // receive/transmit holding reg. / divisor latch LSB
	sc16is750RegIerDlh                               // interrupt enable reg. / divisor latch MSB
	sc16is750RegIirFcrEfr                            // interrupt identification reg. / FIFO control / enhanced feature reg.
	sc16is750RegLcr                                  // line control reg.
	sc16is750RegMcrXon1                              // modem control reg. / XON1 word
	sc16is750RegLsrXon2                              // line status reg. / XON2 word
	sc16is750RegMsrTcrXoff1                          // modem status reg. / transmission control reg. / XOFF1 word
	sc16is750RegSprTlrXoff2                          // Scratchpad reg. / trigger level reg. / XOFF2 word
	sc16is750RegTxlvl                                // transmit FIFO level reg.
	sc16is750RegRxlvl                                // receive FIFO level reg.
	sc16is750RegIodir                                // I/O pin direction reg.
	sc16is750RegIostate                              // I/O pin states reg.
	sc16is750RegIointena                             // I/O interrupt enable reg.
	sc16is750RegReserved                             // -
	sc16is750RegIocontrol                            // I/O pins control reg.
	sc16is750RegEfcr                                 // extra features reg.
)

// sc16is750Driver is a Gobot Driver for the sc16is750 I2C/SPI to UART, 8-bit I/O, IrDA SIR with 2 address program pins.
// The driver supports also SC16IS760 (higher clock rate possible) and SC16IS740 (no GPIO)
type sc16is750Driver struct {
	name       string
	connector  Connector
	connection Connection
	Config
	gobot.Commander
}

// NewSC16IS750Driver creates a new driver with specified i2c interface
// Params:
//		conn Connector - the Adaptor to use with this Driver
//
// Optional params:
//		i2c.WithBus(int):	bus to use with this driver
//		i2c.WithAddress(int):	address to use with this driver
//
func NewSC16IS750Driver(a Connector, options ...func(Config)) *sc16is750Driver {
	p := &sc16is750Driver{
		name:      gobot.DefaultName("SC16IS750"),
		connector: a,
		Config:    NewConfig(),
		Commander: gobot.NewCommander(),
	}

	for _, option := range options {
		option(p)
	}

	// API commands
	p.AddCommand("WriteGPIO", func(params map[string]interface{}) interface{} {
		pin := params["pin"].(uint8)
		val := params["val"].(uint8)
		err := p.WriteGPIO(pin, val)
		return map[string]interface{}{"err": err}
	})

	p.AddCommand("ReadGPIO", func(params map[string]interface{}) interface{} {
		pin := params["pin"].(uint8)
		val, err := p.ReadGPIO(pin)
		return map[string]interface{}{"val": val, "err": err}
	})

	return p
}

// Name returns the Name for the Driver
func (p *sc16is750Driver) Name() string { return p.name }

// SetName sets the Name for the Driver
func (p *sc16is750Driver) SetName(n string) { p.name = n }

// Connection returns the connection for the Driver
func (p *sc16is750Driver) Connection() gobot.Connection { return p.connector.(gobot.Connection) }

// Start initializes the sc16is750
func (p *sc16is750Driver) Start() (err error) {
	bus := p.GetBusOrDefault(p.connector.GetDefaultBus())
	addressGPIO := p.GetAddressOrDefault(sc16is750Address)
	p.connection, err = p.connector.GetConnection(addressGPIO, bus)
	if err != nil {
		return err
	}

	return
}

// Halt stops the device
func (p *sc16is750Driver) Halt() (err error) { return }

// WriteGPIO writes a value to a gpio pin (0-7)
func (p *sc16is750Driver) WriteGPIO(pin uint8, val uint8) (err error) {
	// read current value of port
	cVal, err := p.readRegister(sc16is750RegIostate)
	if err != nil {
		return err
	}
	// set or reset the bit in value
	var nVal uint8
	if val == 0 {
		nVal = sc16is750clearBit(cVal, uint8(pin))
	} else {
		nVal = sc16is750setBit(cVal, uint8(pin))
	}
	// write new value to port
	err = p.writeRegister(sc16is750RegIostate, uint8(nVal))
	if err != nil {
		return err
	}
	return nil
}

// ReadGPIO reads a value from a given gpio pin (0-7)
func (p *sc16is750Driver) ReadGPIO(pin uint8) (val uint8, err error) {
	// read current value of direction register, 0 is no output, 1 is an output
	iodir, err := p.readRegister(sc16is750RegIodir)
	if err != nil {
		return val, err
	}
	// set pin as input by clearing bit
	iodirVal := sc16is750clearBit(iodir, uint8(pin))
	// write direction register
	err = p.writeRegister(sc16is750RegIodir, uint8(iodirVal))
	if err != nil {
		return val, err
	}
	// read port and create return bit
	val, err = p.readRegister(sc16is750RegIostate)
	if err != nil {
		return val, err
	}
	val = 1 << uint8(pin) & val
	if val > 1 {
		val = 1
	}
	return val, nil
}

func sc16is750setBit(n uint8, pos uint8) uint8 {
	n |= (1 << pos)
	return n
}

func sc16is750clearBit(n uint8, pos uint8) uint8 {
	mask := ^uint8(1 << pos)
	n &= mask
	return n
}

// write the content of the given register
func (p *sc16is750Driver) writeRegister(regAddress sc16is750Register, val uint8) error {
	// write content of requested register
	return p.connection.WriteByteData(uint8(regAddress), val)
}

// read the content of the given register
func (p *sc16is750Driver) readRegister(regAddress sc16is750Register) (uint8, error) {
	return p.connection.ReadByteData(uint8(regAddress))
}
