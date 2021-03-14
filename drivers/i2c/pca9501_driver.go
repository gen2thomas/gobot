package i2c

import (
	"fmt"
	"gobot.io/x/gobot"
	"log"
)

// PCA9501 supports addresses from 0x00 to 0x7F
// 0x00 - 0x3F: GPIO
// 0x40 - 0x7F: EEPROM
// Example: 0x04 GPIO, 0x44 is EEPROM
const pca9501AddressGPIO = 0x04

// Set bit 0x40 in connection address will activate EEPROM access
const pca9501AddressMem = pca9501AddressGPIO | 0x40

var pca9501Debug = false // will be more verbose when set to true

// PCA9501Driver is a Gobot Driver for the PCA9501 8-bit GPIO  & 2-kbit EEPROM with 6 address program pins.
// 0 EE A5 A4 A3 A2 A1 A0|rd
// Lowest bit (rd) is mapped to switch between write(0)/read(1), it is not part of the "real" address.
// Highest bit (EE) is mapped to switch between GPIO(0)/EEPROM(1).
//
// Example: A1,A2=1, others are 0
// Address mask => 1000110|1 => real 7-bit address mask 0100 0110 = 0x46
//
// 2-kbit EEPROM has 250 byte, means addresses between 0x00-0xFA
//
type PCA9501Driver struct {
	name           string
	connector      Connector
	connectionGPIO Connection
	connectionMem  Connection
	Config
	gobot.Commander
}

// NewPCA9501Driver creates a new driver with specified i2c interface
// Params:
//		conn Connector - the Adaptor to use with this Driver
//
// Optional params:
//		i2c.WithBus(int):	bus to use with this driver
//		i2c.WithAddress(int):	address to use with this driver
//
func NewPCA9501Driver(a Connector, options ...func(Config)) *PCA9501Driver {
	p := &PCA9501Driver{
		name:      gobot.DefaultName("PCA9501"),
		connector: a,
		Config:    NewConfig(),
		Commander: gobot.NewCommander(),
	}

	for _, option := range options {
		option(p)
	}

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
func (p *PCA9501Driver) Name() string { return p.name }

// SetName sets the Name for the Driver
func (p *PCA9501Driver) SetName(n string) { p.name = n }

// Connection returns the connection for the Driver
func (p *PCA9501Driver) Connection() gobot.Connection { return p.connector.(gobot.Connection) }

// Start initializes the pca9501
func (p *PCA9501Driver) Start() (err error) {
	bus := p.GetBusOrDefault(p.connector.GetDefaultBus())
	addressGPIO := p.GetAddressOrDefault(pca9501AddressGPIO)
	p.connectionGPIO, err = p.connector.GetConnection(addressGPIO, bus)
	if err != nil {
		return err
	}
	addressMem := p.GetAddressOrDefault(pca9501AddressMem)
	p.connectionMem, err = p.connector.GetConnection(addressMem, bus)
	if err != nil {
		return err
	}

	return
}

// Halt stops the device
func (p *PCA9501Driver) Halt() (err error) { return }

// WriteGPIO writes a value to a gpio pin (0-7)
func (p *PCA9501Driver) WriteGPIO(pin uint8, val uint8) (err error) {
	// read current value of CTRL register, 0 is no output, 1 is an output
	iodir, err := p.read()
	if err != nil {
		return err
	}
	// set pin as output by clearing bit
	iodirVal := clearBitAtPos(iodir, uint8(pin))
	// write CTRL register
	err = p.write(uint8(iodirVal))
	if err != nil {
		return err
	}
	// read current value of port
	cVal, err := p.read()
	if err != nil {
		return err
	}
	// set or reset the bit in value
	var nVal uint8
	if val == 0 {
		nVal = clearBitAtPos(cVal, uint8(pin))
	} else {
		nVal = setBitAtPos(cVal, uint8(pin))
	}
	// write new value to port
	err = p.write(uint8(nVal))
	if err != nil {
		return err
	}
	return nil
}

// ReadGPIO reads a value from a given gpio pin (0-7)
func (p *PCA9501Driver) ReadGPIO(pin uint8) (val uint8, err error) {
	// read current value of CTRL register, 0 is no output, 1 is an output
	iodir, err := p.read()
	if err != nil {
		return 0, err
	}
	// set pin as input by setting bit
	iodirVal := setBitAtPos(iodir, uint8(pin))
	// write CTRL register
	err = p.write(uint8(iodirVal))
	if err != nil {
		return 0, err
	}
	// read port and create return bit
	val, err = p.read()
	if err != nil {
		return val, err
	}
	val = 1 << uint8(pin) & val
	if val > 1 {
		val = 1
	}
	return val, nil
}

// ReadEEPROM reads a value from a given address (00-FA)
func (p *PCA9501Driver) ReadEEPROM(address uint8) (val uint8, err error) {
	// write EEPROM address to read from
	err = p.writemem(uint8(address))
	if err != nil {
		return 0, err
	}
	// read value
	val, err = p.readmem()
	return val, err
}

// WriteEEPROM writes a value to a given address im memory (00-FA)
func (p *PCA9501Driver) WriteEEPROM(address uint8, val uint8) (err error) {
	// write EEPROM address to write to
	err = p.writemem(uint8(address))
	if err != nil {
		return err
	}
	// write new value to port
	err = p.writemem(uint8(val))
	if err != nil {
		return err
	}
	return nil
}

// write the given value to the GPIO connection
func (p *PCA9501Driver) write(val uint8) (err error) {
	if pca9501Debug {
		log.Printf("write: PCA9501 address: 0x%X, value: 0x%X\n", p.GetAddressOrDefault(pca9501AddressGPIO), val)
	}
	if _, err = p.connectionGPIO.Write([]uint8{val}); err != nil {
		return err
	}
	return nil
}

// write the given value to the memory connection
func (p *PCA9501Driver) writemem(val uint8) (err error) {
	if pca9501Debug {
		log.Printf("write: PCA9501 address: 0x%X, value: 0x%X\n", p.GetAddressOrDefault(pca9501AddressMem), val)
	}
	if _, err = p.connectionMem.Write([]uint8{val}); err != nil {
		return err
	}
	return nil
}

// read get the value from the GPIO connection
func (p *PCA9501Driver) read() (val uint8, err error) {
	buf := []byte{0}
	bytesRead, err := p.connectionGPIO.Read(buf)
	if err != nil {
		return val, err
	}
	if bytesRead != 1 {
		err = ErrNotEnoughBytes
		return
	}
	if pca9501Debug {
		log.Printf("reading: PCA9501 address: 0x%X, value: 0x%X\n", p.GetAddressOrDefault(pca9501AddressGPIO), buf)
	}
	return buf[0], nil
}

// read get the value from the memory connection
func (p *PCA9501Driver) readmem() (val uint8, err error) {
	buf := []byte{0}
	bytesRead, err := p.connectionMem.Read(buf)
	if err != nil {
		return val, err
	}
	if bytesRead != 1 {
		err = ErrNotEnoughBytes
		return
	}
	if pca9501Debug {
		log.Printf("reading: PCA9501 address: 0x%X, value: 0x%X\n", p.GetAddressOrDefault(pca9501AddressMem), buf)
	}
	return buf[0], nil
}

func setBitAtPos(n uint8, pos uint8) uint8 {
	n |= (1 << pos)
	return n
}

func clearBitAtPos(n uint8, pos uint8) uint8 {
	mask := ^uint8(1 << pos)
	n &= mask
	return n
}
