package i2c

import (
	"gobot.io/x/gobot"
)

// sc16is750 has 2 address pins and supports 16 addresses from 0x48 to 0x57
// by connecting pins to VDD, VSS, SCL, SDA
// beside this hardware address definition it supports subaddresses by I2C command
// A connection is mandatory, because there are no internal resistors.
// Default address 0x4D is for both pins are grounded.
const sc16is750Address = 0x4D

// sc16is750Driver is a Gobot Driver for the sc16is750 I2C/SPI to UART, 8-bit I/O, IrDA SIR with 2 address program pins.
// The driver supports also SC16IS760 (higher clock rate possible) and SC16IS740 (no GPIO)
type sc16is750Driver struct {
	name           string
	connector      Connector
	connectionGPIO Connection
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
	p.connectionGPIO, err = p.connector.GetConnection(addressGPIO, bus)
	if err != nil {
		return err
	}

	return
}

// Halt stops the device
func (p *sc16is750Driver) Halt() (err error) { return }

// WriteGPIO writes a value to a gpio pin (0-7)
func (p *sc16is750Driver) WriteGPIO(pin uint8, val uint8) (err error) {
	// read current value of CTRL register, 0 is no output, 1 is an output
	iodir, err := p.connectionGPIO.ReadByte()
	if err != nil {
		return err
	}
	// set pin as output by clearing bit
	iodirVal := clearBitAtPosition(iodir, uint8(pin))
	// write CTRL register
	err = p.connectionGPIO.WriteByte(uint8(iodirVal))
	if err != nil {
		return err
	}
	// read current value of port
	cVal, err := p.connectionGPIO.ReadByte()
	if err != nil {
		return err
	}
	// set or reset the bit in value
	var nVal uint8
	if val == 0 {
		nVal = clearBitAtPosition(cVal, uint8(pin))
	} else {
		nVal = setBitAtPosition(cVal, uint8(pin))
	}
	// write new value to port
	err = p.connectionGPIO.WriteByte(uint8(nVal))
	if err != nil {
		return err
	}
	return nil
}

// ReadGPIO reads a value from a given gpio pin (0-7)
func (p *sc16is750Driver) ReadGPIO(pin uint8) (val uint8, err error) {
	// read current value of CTRL register, 0 is no output, 1 is an output
	iodir, err := p.connectionGPIO.ReadByte()
	if err != nil {
		return 0, err
	}
	// set pin as input by setting bit
	iodirVal := setBitAtPosition(iodir, uint8(pin))
	// write CTRL register
	err = p.connectionGPIO.WriteByte(uint8(iodirVal))
	if err != nil {
		return 0, err
	}
	// read port and create return bit
	val, err = p.connectionGPIO.ReadByte()
	if err != nil {
		return val, err
	}
	val = 1 << uint8(pin) & val
	if val > 1 {
		val = 1
	}
	return val, nil
}

func setBitAtPosition(n uint8, pos uint8) uint8 {
	n |= (1 << pos)
	return n
}

func clearBitAtPosition(n uint8, pos uint8) uint8 {
	mask := ^uint8(1 << pos)
	n &= mask
	return n
}
