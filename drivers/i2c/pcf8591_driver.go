package i2c

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gobot.io/x/gobot"
)

// PCF8591 supports addresses from 0x48 to 0x4F
// The default address applies when all address pins connected to ground.
const pcf8591DefaultAddress = 0x48

const pcf8591Debug = true

type mode uint8
type channel uint8

const (
	pcf8591_CHAN0 channel = 0x00
	pcf8591_CHAN1         = 0x01
	pcf8591_CHAN2         = 0x02
	pcf8591_CHAN3         = 0x03
)

const pcf8591_AION = 0x04 // auto increment, only relevant for ADC

const (
	pcf8591_ALLSINGLE mode = 0x00
	pcf8591_THREEDIFF      = 0x10
	pcf8591_MIXED          = 0x20
	pcf8591_TWODIFF        = 0x30
	pcf8591_ANAON          = 0x40
)

const pcf8591_ADMASK = 0x33 // channels and mode

type modeChan struct {
	mode    mode
	channel channel
}

// modeMap is to define the relation between a given description and the mode and channel
// beside the long form there are some short forms available without risk of confusion
//
// pure single mode
// "s.0"..."s.3": read single value of input n => channel n
// pure differential mode
// "d.0-1": differential value between input 0 and 1 => channel 0
// "d.2-3": differential value between input 2 and 3 => channel 1
// mixed mode
// "m.0": single value of input 0  => channel 0
// "m.1": single value of input 1  => channel 1
// "m.2-3": differential value between input 2 and 3 => channel 2
// three differential inputs, related to input 3
// "t.0-3": differential value between input 0 and 3 => channel 0
// "t.1-3": differential value between input 1 and 3 => channel 1
// "t.2-3": differential value between input 1 and 3 => channel 2
var modeMap = map[string]modeChan{
	"s.0":   {pcf8591_ALLSINGLE, pcf8591_CHAN0},
	"0":     {pcf8591_ALLSINGLE, pcf8591_CHAN0},
	"s.1":   {pcf8591_ALLSINGLE, pcf8591_CHAN1},
	"1":     {pcf8591_ALLSINGLE, pcf8591_CHAN1},
	"s.2":   {pcf8591_ALLSINGLE, pcf8591_CHAN2},
	"2":     {pcf8591_ALLSINGLE, pcf8591_CHAN2},
	"s.3":   {pcf8591_ALLSINGLE, pcf8591_CHAN3},
	"3":     {pcf8591_ALLSINGLE, pcf8591_CHAN3},
	"d.0-1": {pcf8591_TWODIFF, pcf8591_CHAN0},
	"0-1":   {pcf8591_TWODIFF, pcf8591_CHAN0},
	"d.2-3": {pcf8591_TWODIFF, pcf8591_CHAN1},
	"m.0":   {pcf8591_MIXED, pcf8591_CHAN0},
	"m.1":   {pcf8591_MIXED, pcf8591_CHAN1},
	"m.2-3": {pcf8591_MIXED, pcf8591_CHAN2},
	"t.0-3": {pcf8591_THREEDIFF, pcf8591_CHAN0},
	"0-3":   {pcf8591_THREEDIFF, pcf8591_CHAN0},
	"t.1-3": {pcf8591_THREEDIFF, pcf8591_CHAN1},
	"1-3":   {pcf8591_THREEDIFF, pcf8591_CHAN1},
	"t.2-3": {pcf8591_THREEDIFF, pcf8591_CHAN2},
}

const rotateCount = 0 // usefully for debugging purposes
const skipReads = 3

// PCF8591Driver is a Gobot Driver for the PCF8591 8-bit 4xA/D & 1xD/A converter with i2c interface and 3 address pins.
// The analog inputs can be used as differential inputs in different ways.
//
// Address specification:
// 1 0 0 1 A2 A1 A0|rd
// Lowest bit (rd) is mapped to switch between write(0)/read(1), it is not part of the "real" address.
//
// Example: A1,A2=1, others are 0
// Address mask => 1001110|1 => real 7-bit address mask 0100 1110 = 0x4E
//
// For example, here is the Adafruit board that uses this chip:
// https://www.adafruit.com/product/4648
//
// This driver was tested with Tinkerboard and this board with temperature & brightness sensor:
// https://www.makershop.de/download/YL_40_PCF8591.pdf
//
type PCF8591Driver struct {
	name       string
	connector  Connector
	connection Connection
	Config
	gobot.Commander
	lastCtrlByte uint8
	lastAnaOut   uint8
	LastRead     [rotateCount*4 + 1][]byte
}

// NewPCF8591Driver creates a new driver with specified i2c interface
// Params:
//		conn Connector - the Adaptor to use with this Driver
//
// Optional params:
//		i2c.WithBus(int):	bus to use with this driver
//		i2c.WithAddress(int):	address to use with this driver
//
func NewPCF8591Driver(a Connector, options ...func(Config)) *PCF8591Driver {
	p := &PCF8591Driver{
		name:      gobot.DefaultName("PCF8591"),
		connector: a,
		Config:    NewConfig(),
		Commander: gobot.NewCommander(),
	}

	return p
}

// Name returns the Name for the Driver
func (p *PCF8591Driver) Name() string { return p.name }

// SetName sets the Name for the Driver
func (p *PCF8591Driver) SetName(n string) { p.name = n }

// Connection returns the connection for the Driver
func (p *PCF8591Driver) Connection() gobot.Connection { return p.connector.(gobot.Connection) }

// Start initializes the PCF8591
func (p *PCF8591Driver) Start() (err error) {
	bus := p.GetBusOrDefault(p.connector.GetDefaultBus())
	address := p.GetAddressOrDefault(pcf8591DefaultAddress)

	p.connection, err = p.connector.GetConnection(address, bus)
	if err != nil {
		return err
	}

	if err := p.AnalogOutputState(false); err != nil {
		return err
	}
	return
}

// Halt stops the device
func (p *PCF8591Driver) Halt() (err error) {
	return p.AnalogOutputState(false)
}

// AnalogRead returns value from analog reading of given input description
//
// Vlsb = (Vref-Vagnd)/256
// values are related to Vlsb by (Van+ - Van-)/Vlsb, Van-=0 for single mode
//
// The first read contains the last converted value (usually the last read).
// After the channel was switched this means the value of the previous read channel.
// After power on, the first byte read will be 80h, because the read is one cycle behind.
//
// Important note:
// With a bus speed of 100 kBit/sec, the ADC conversion has ~80 us + ACK (time to transfer the previous value).
// This time seems to be a little bit to small (datasheet 90 us).
// A active bus driver don't fix it (it seems rather the opposite).
//
// This leads to following behavior:
// * the transition process takes an additional cycle, very often
// * some circuits takes one cycle longer transition time in addition
// * reading more than one byte by Read([]byte), e.g. to calculate an average, is not sufficient,
//   because some missing integration steps in each byte (each byte is a little bit to small)
//
// So, for default, we drop the first three bytes to get the right value.
func (p *PCF8591Driver) AnalogRead(description string) (value int, err error) {
	mc, err := pcf8591GetModeChannel(description)
	if err != nil {
		return 0, err
	}

	// reset channel and mode
	ctrlByte := p.lastCtrlByte & ^uint8(pcf8591_ADMASK)
	// set to current channel and mode, AI must be off, because we need reading twice
	ctrlByte = ctrlByte | uint8(mc.mode) | uint8(mc.channel) & ^uint8(pcf8591_AION)
	if err = p.writeCtrlByte(ctrlByte); err != nil {
		return 0, err
	}

	// initiate read but skip some bytes
	for i := uint8(0); i < rotateCount*4+1; i++ {
		p.readBuf(i)
	}

	// additional relax time
	time.Sleep(1 * time.Millisecond)

	// real read
	uval, err := p.connection.ReadByte()
	if err != nil {
		return 0, err
	}

	value = int(uval)
	if mc.pcf8591IsDiff() {
		value = value - 128
	}

	return value, err
}

func (p *PCF8591Driver) readBuf(nr uint8) error {
	buf := make([]byte, skipReads)
	cntRead, err := p.connection.Read(buf)
	if err != nil {
		return err
	}
	if cntRead != len(buf) {
		return fmt.Errorf("Not enough bytes (%d of %d) read", cntRead, len(buf))
	}
	if pcf8591Debug {
		p.LastRead[nr] = buf
	}
	return nil
}

// AnalogWrite writes the given value to the analog output (DAC)
// Vlsb = (Vref-Vagnd)/256
// Vaout = Vagnd+Vlsb*value
func (p *PCF8591Driver) AnalogWrite(value uint8) (err error) {
	if p.lastAnaOut == value {
		if pcf8591Debug {
			log.Printf("write skipped because value unchanged: 0x%X\n", value)
		}
		return nil
	}

	ctrlByte := p.lastCtrlByte | pcf8591_ANAON
	cntWritten, err := p.connection.Write([]byte{ctrlByte, value})
	if err != nil {
		return err
	}

	if cntWritten != 2 {
		return fmt.Errorf("Not enough bytes (%d of %d) written", cntWritten, 2)
	}

	p.lastCtrlByte = ctrlByte
	p.lastAnaOut = value
	return nil
}

// AnalogOutputState enables or disables the analog output
// Please note that in case of using the internal oscillator
// and the auto increment mode the output should not switched off.
// Otherwise conversion errors could occur.
func (p *PCF8591Driver) AnalogOutputState(state bool) (err error) {
	var ctrlByte uint8
	if state {
		ctrlByte = p.lastCtrlByte | pcf8591_ANAON
	} else {
		ctrlByte = p.lastCtrlByte & ^uint8(pcf8591_ANAON)
	}

	if err = p.writeCtrlByte(ctrlByte); err != nil {
		return err
	}
	return nil
}

func (p *PCF8591Driver) writeCtrlByte(ctrlByte uint8) error {
	if p.lastCtrlByte != ctrlByte {
		if err := p.connection.WriteByte(ctrlByte); err != nil {
			return err
		}
		p.lastCtrlByte = ctrlByte
	} else {
		if pcf8591Debug {
			log.Printf("write skipped because control byte unchanged: 0x%X\n", ctrlByte)
		}
	}
	return nil
}

func pcf8591GetModeChannel(description string) (*modeChan, error) {
	mc, ok := modeMap[description]
	if !ok {
		descriptions := []string{}
		for k := range modeMap {
			descriptions = append(descriptions, k)
		}
		ds := strings.Join(descriptions, ", ")
		return nil, fmt.Errorf("Unknown description '%s' for read analog value, accepted values: %s", description, ds)
	}

	return &mc, nil
}

func (mc modeChan) pcf8591IsDiff() bool {
	switch mc.mode {
	case pcf8591_TWODIFF:
		return true
	case pcf8591_THREEDIFF:
		return true
	case pcf8591_MIXED:
		return mc.channel == pcf8591_CHAN2
	default:
		return false
	}
}
