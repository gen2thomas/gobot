package i2c

import (
	"fmt"
	"gobot.io/x/gobot"
	"time"
)

// PCA9501 supports addresses from 0x00 to 0x7F
// 0x00 - 0x3F: GPIO
// 0x40 - 0x7F: EEPROM
// Example: 0x04 GPIO, 0x44 is EEPROM
const pca9501Address = 0x04

// This EEPROM address (range 0x00-0xFE) is not usable for other meaningfull r/w-operations
// Please read explanation at the end of this document
const pca9501MemReadDummyAddress = 0x00

// Value does not matter, could be used to identify the dummy address, when unique
// Please read explanation at the end of this document
const pca9501MemReadDummyValue = 0x15

// PCA9501Driver is a Gobot Driver for the PCA9501 8-bit GPIO  & 2-kbit EEPROM with 6 address program pins.
// 0 EE A5 A4 A3 A2 A1 A0|rd
// Lowest bit (rd) is mapped to switch between write(0)/read(1), it is not part of the "real" address.
// Highest bit (EE) is mapped to switch between GPIO(0)/EEPROM(1).
//
// Example: A1,A2=1, others are 0
// Address mask => 1000110|1 => real 7-bit address mask 0100 0110 = 0x46
//
// 2-kbit EEPROM has 256 byte, means addresses between 0x00-0xFF
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

	p.AddCommand("WriteEEPROM", func(params map[string]interface{}) interface{} {
		address := params["address"].(uint8)
		val := params["val"].(uint8)
		err := p.WriteEEPROM(address, val)
		return map[string]interface{}{"err": err}
	})

	p.AddCommand("ReadEEPROM", func(params map[string]interface{}) interface{} {
		address := params["address"].(uint8)
		val, err := p.ReadEEPROM(address)
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
	addressGPIO := p.GetAddressOrDefault(pca9501Address)
	p.connectionGPIO, err = p.connector.GetConnection(addressGPIO, bus)
	if err != nil {
		return err
	}
	addressMem := p.getAddressMem(pca9501Address)
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
	iodir, err := p.connectionGPIO.ReadByte()
	if err != nil {
		return err
	}
	// set pin as output by clearing bit
	iodirVal := clearBitAtPos(iodir, uint8(pin))
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
		nVal = clearBitAtPos(cVal, uint8(pin))
	} else {
		nVal = setBitAtPos(cVal, uint8(pin))
	}
	// write new value to port
	err = p.connectionGPIO.WriteByte(uint8(nVal))
	if err != nil {
		return err
	}
	return nil
}

// ReadGPIO reads a value from a given gpio pin (0-7)
func (p *PCA9501Driver) ReadGPIO(pin uint8) (val uint8, err error) {
	// read current value of CTRL register, 0 is no output, 1 is an output
	iodir, err := p.connectionGPIO.ReadByte()
	if err != nil {
		return 0, err
	}
	// set pin as input by setting bit
	iodirVal := setBitAtPos(iodir, uint8(pin))
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

// ReadEEPROM reads a value from a given address (0x00-0xFF)
func (p *PCA9501Driver) ReadEEPROM(address uint8) (val uint8, err error) {
	// Please read explanation at the end of this document to understand, why it is implemented in this way
	if address == pca9501MemReadDummyAddress {
		return pca9501MemReadDummyValue, fmt.Errorf("Dummy address %d not meaningfull to read\n", pca9501MemReadDummyAddress)
	}
	// write dummy value to set the address counter to n
	err = p.connectionMem.WriteByteData(pca9501MemReadDummyAddress, pca9501MemReadDummyValue)
	if err != nil {
		return val, err
	}
	time.Sleep(10 * time.Millisecond)
	// read all addresses, starting with n+1
	buf := make([]uint8, 255)
	_, err = p.connectionMem.Read(buf)
	if err != nil {
		return val, err
	}
	return buf[address-1-pca9501MemReadDummyAddress], err
}

// WriteEEPROM writes a value to a given address in memory (0x00-0xFF)
func (p *PCA9501Driver) WriteEEPROM(address uint8, val uint8) (err error) {
	if address == pca9501MemReadDummyAddress {
		return fmt.Errorf("Dummy address %d not meaningfull to write\n", pca9501MemReadDummyAddress)
	}
	return p.connectionMem.WriteByteData(address, val)
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

func (p *PCA9501Driver) getAddressMem(defaultAdress int) int {
	return p.GetAddressOrDefault(defaultAdress) | 0x40
}

// Explanation for EEPROM read possibilities of PCA9501 and ristrictions with adaptor implementations
// PCA9501 has an internal address counter "n" and supports 3 EEPROM read methodes
// * read value of position "n" (current address read)
// * set address counter "n" by a write-counter-operation, then read (random address read)
// * read all 255 bytes, starting with "n" (sequential read)
//
// for further reading:
// * STARTW - Start condition with device address and write
// * STARTR - Start condition with device address and read
//
// The most usable feature seems to be the "random address read":
// According to specification we have to implement a sequence of "STARTW-DATA1-STARTR-DATA2-STOP"
// DATA1: EEPROM-address to set (will set "n"), DATA2: value at "n" read
// This sequence (with missing STOP after DATA1) is not supported by (some) implementations of gobot adaptors.
//
// Some words to "current address read":
// After an reset of device ("n" should be 0) and some write operations the current address "n" is unknown, except
// we would have an local counter reflecting the state of "n". The device don't provide "n" in any way.
// Therefore "current address read" is hard to implement in a safe way.
//
// Using "sequential read" to fullfill gobot adaptors implementations and provide a "pseudo random read":
// After a write operation to the address "n" the counter will be incremented to "n+1". So, it will be always possible
// to read, which will return the value of address "n+1" (see "current address read").
// To make our own "pseudo random read", we have to use a dummy address to write to and afterwards read complete
// memory to a buffer. Because we know "n" at the moment of writing, we can match all buffer elements to a EEPROM address
//
// For this make to work there are 2 additional constants defined
//
// "pca9501MemReadDummyAddress":
// We have to use an dummy write (start-write-stop) to this address to set the EEPROM address
// counter "n" for sequential read (start-read-stop) afterwards, beginning at "n+1"
// It is possible to use each address in range 0x00-0xFE for dummy write, this also means
// the chosen EEPROM address is not usable for other meaningfull r/w-operations!
// Please note that 0xFF will not work for unknown reason.
//
// "pca9501MemReadDummyValue"
// Value does not matter, could be used to identify the dummy address, when unique in your application.
