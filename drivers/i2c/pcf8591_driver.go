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

const (
	pcf8591Debug       = true
	pcf8591RotateCount = 0 // usefully for debugging purposes
)

type pcf8591Mode uint8
type PCF8591Channel uint8

const (
	pcf8591_CHAN0 PCF8591Channel = 0x00
	pcf8591_CHAN1                = 0x01
	pcf8591_CHAN2                = 0x02
	pcf8591_CHAN3                = 0x03
)

const pcf8591_AION = 0x04 // auto increment, only relevant for ADC

const (
	pcf8591_ALLSINGLE pcf8591Mode = 0x00
	pcf8591_THREEDIFF             = 0x10
	pcf8591_MIXED                 = 0x20
	pcf8591_TWODIFF               = 0x30
	pcf8591_ANAON                 = 0x40
)

const pcf8591_ADMASK = 0x33 // channels and mode

type pcf8591ModeChan struct {
	mode    pcf8591Mode
	channel PCF8591Channel
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
var pcf8591ModeMap = map[string]pcf8591ModeChan{
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

// PCF8591Driver is a Gobot Driver for the PCF8591 8-bit 4xA/D & 1xD/A converter with i2c interface and 3 address pins.
// The analog inputs can be used as differential inputs in different ways.
//
// All values are linear scaled to 3.3V by default. This can be changed, see example "tinkerboard_pcf8591.go".
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
	lastCtrlByte   byte
	lastAnaOut     byte
	additionalSkip uint8
	toMax          [4]int
	toMin          [4]int
	fromMin        int
	fromMax        int
	rescaleAI      func(value byte, mc pcf8591ModeChan) (milliVolt int)
	rescaleAO      func(milliVolt int) (value byte)
	LastRead       [pcf8591RotateCount*4 + 1][]byte // for debugging purposes
}

// NewPCF8591Driver creates a new driver with specified i2c interface
// Params:
//    conn Connector - the Adaptor to use with this Driver
//
// Optional params:
//    i2c.WithBus(int): bus to use with this driver
//    i2c.WithAddress(int): address to use with this driver
//    i2c.WithPCF8591AdditionalSkip(uint8): additional skip value to stabilize read
//    i2c.WithPCF8591RescaleInput(uint8, int, int): set scale values for AI
//    i2c.WithPCF8591RescaleOutput(int, int): set scale values for AO
//
func NewPCF8591Driver(a Connector, options ...func(Config)) *PCF8591Driver {
	p := &PCF8591Driver{
		name:           gobot.DefaultName("PCF8591"),
		connector:      a,
		Config:         NewConfig(),
		Commander:      gobot.NewCommander(),
		additionalSkip: 2,
	}

	// in case of 32 bit int => -2147483648 to 2147483647, voltage in mV (1V = 1000mV)
	for i := 0; i < 4; i++ {
		p.toMin[i] = 0
		p.toMax[i] = 3300
	}

	p.fromMin = 0
	p.fromMax = 3300

	p.rescaleAI = func(value byte, mc pcf8591ModeChan) (milliVolt int) {
		// return in milliVolt only for the default
		fromMin := 0
		fromMax := 255
		iVal := int(value)
		if mc.pcf8591IsDiff() {
			if iVal > 127 {
				// first bit is set, means negative
				iVal = iVal - 256
			}
			fromMin = -128
			fromMax = 127
		}
		return pcf8591Rescale(iVal, fromMin, fromMax, int(p.toMin[mc.channel]), int(p.toMax[mc.channel]))
	}

	p.rescaleAO = func(milliVolt int) byte {
		// given value in milliVolt only for the default
		return byte(pcf8591Rescale(milliVolt, p.fromMin, p.fromMax, 0, 255))
	}

	for _, option := range options {
		option(p)
	}

	return p
}

// WithPCF8591AdditionalSkip option sets the PCF8591 additionalSkip value
func WithPCF8591AdditionalSkip(val uint8) func(Config) {
	return func(c Config) {
		p, ok := c.(*PCF8591Driver)
		if ok {
			p.additionalSkip = val
		} else {
			panic("trying to set skipping additional input bytes value for non-PCF8591Driver")
		}
	}
}

// WithPCF8591RescaleInput option sets the PCF8591 scale values, toMin and toMax value for the given input channel
func WithPCF8591RescaleInput(channel PCF8591Channel, toMin, toMax int) func(Config) {
	return func(c Config) {
		p, ok := c.(*PCF8591Driver)
		if ok {
			p.toMin[channel] = toMin
			p.toMax[channel] = toMax
		} else {
			panic("trying to set input scale values for non-PCF8591Driver")
		}
	}
}

// WithPCF8591RescaleOutput option sets the PCF8591 scale values, fromMin and fromMax value for the analog output
func WithPCF8591RescaleOutput(fromMin, fromMax int) func(Config) {
	return func(c Config) {
		p, ok := c.(*PCF8591Driver)
		if ok {
			p.fromMin = fromMin
			p.fromMax = fromMax
		} else {
			panic("trying to set output scale values for non-PCF8591Driver")
		}
	}
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
// Vlsb = (Vref-Vagnd)/256, value = (Van+ - Van-)/Vlsb, Van-=Vagnd for single mode
//
// The first read contains the last converted value (usually the last read).
// After the channel was switched this means the value of the previous read channel.
// After power on, the first byte read will be 80h, because the read is one cycle behind.
//
// Important note:
// With a bus speed of 100 kBit/sec, the ADC conversion has ~80 us + ACK (time to transfer the previous value).
// This time seems to be a little bit to small (datasheet 90 us).
// An i2c bus extender (LTC4311) don't fix it (it seems rather the opposite).
//
// This leads to following behavior:
// * the transition process takes an additional cycle, very often
// * some circuits takes one cycle longer transition time in addition
// * reading more than one byte by Read([]byte), e.g. to calculate an average, is not sufficient,
//   because some missing integration steps in each conversion (each byte value is a little bit lower than expected)
//
// So, for default, we drop the first three bytes to get the right value.
func (p *PCF8591Driver) AnalogRead(description string) (value int, err error) {
	mc, err := PCF8591ParseModeChan(description)
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
	for i := uint8(0); i < pcf8591RotateCount*4+1; i++ {
		if err := p.readBuf(i, 1+p.additionalSkip); err != nil {
			return 0, err
		}
	}

	// additional relax time
	time.Sleep(1 * time.Millisecond)

	// real used read
	uval, err := p.connection.ReadByte()
	if err != nil {
		return 0, err
	}

	return p.rescaleAI(uval, *mc), err
}

// AnalogWrite writes the given value to the analog output (DAC)
// Vlsb = (Vref-Vagnd)/256, Vaout = Vagnd+Vlsb*value
func (p *PCF8591Driver) AnalogWrite(value int) (err error) {

	byteVal := p.rescaleAO(value)

	if p.lastAnaOut == byteVal {
		if pcf8591Debug {
			log.Printf("write skipped because value unchanged: 0x%X\n", byteVal)
		}
		return nil
	}

	ctrlByte := p.lastCtrlByte | pcf8591_ANAON
	err = p.connection.WriteByteData(ctrlByte, byteVal)
	if err != nil {
		return err
	}

	p.lastCtrlByte = ctrlByte
	p.lastAnaOut = byteVal
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

// PCF8591ParseModeChan is used to get a working combination between mode (single, mixed, 2 differential, 3 differential)
// and the related channel to read from, parsed from the given description string.
func PCF8591ParseModeChan(description string) (*pcf8591ModeChan, error) {
	mc, ok := pcf8591ModeMap[description]
	if !ok {
		descriptions := []string{}
		for k := range pcf8591ModeMap {
			descriptions = append(descriptions, k)
		}
		ds := strings.Join(descriptions, ", ")
		return nil, fmt.Errorf("Unknown description '%s' for read analog value, accepted values: %s", description, ds)
	}

	return &mc, nil
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

func (p *PCF8591Driver) readBuf(nr uint8, cntBytes uint8) error {
	buf := make([]byte, cntBytes)
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

func pcf8591Rescale(input, fromMin, fromMax, toMin, toMax int) int {

	if input < fromMin {
		input = fromMin
	}
	if input > fromMax {
		input = fromMax
	}
	return (input-fromMin)*(toMax-toMin)/(fromMax-fromMin) + toMin
}

func (mc pcf8591ModeChan) pcf8591IsDiff() bool {
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
