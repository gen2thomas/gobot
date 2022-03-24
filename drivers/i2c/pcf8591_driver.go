package i2c

import (
	"fmt"
	"strings"

	"gobot.io/x/gobot"
)

// PCF8591 supports addresses from 0x48 to 0x4F
// The default address applies when all address pins connected to ground.
const pcf8591DefaultAddress = 0x48

const (
	pcf8591_CHANMASK = 0x03
	pcf8591_AIMASK   = 0x04 // only relevant for ADC
	pcf8591_ADMASK   = 0x30
	pcf8591_DAMASK   = 0x40
)

type mode uint8
type channel uint8

const (
	pcf8591_CHAN0 channel = 0x00
	pcf8591_CHAN1         = 0x01
	pcf8591_CHAN2         = 0x02
	pcf8591_CHAN3         = 0x03
)

const pcf8591_AION = 0x04 // auto increment

const (
	pcf8591_ALLSINGLE mode = 0x00
	pcf8591_THREEDIFF      = 0x10
	pcf8591_MIXED          = 0x20
	pcf8591_TWODIFF        = 0x30
	pcf8591_ANAON          = 0x40
)

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
type PCF8591Driver struct {
	name       string
	connector  Connector
	connection Connection
	Config
	gobot.Commander
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
// Vlsb = (Vref-Vagnd)/256
// values are related to Vlsb by (Van+ - Van-)/Vlsb, Van-=0 for single mode
// After power on, the first byte read will be 80h, because the read is one cycle behind.
func (p *PCF8591Driver) AnalogRead(description string) (value int, err error) {
	mc, ok := modeMap[description]
	if !ok {
		descriptions := []string{}
		for k := range modeMap {
			descriptions = append(descriptions, k)
		}
		ds := strings.Join(descriptions, ", ")
		return 0, fmt.Errorf("Unknown description '%s' for read analog value, accepted values: %s", description, ds)
	}

	// ANAON is needed for ADC
	ctrlByte := uint8(pcf8591_ANAON | uint8(mc.mode) | uint8(mc.channel))
	if err = p.connection.WriteByte(ctrlByte); err != nil {
		return 0, err
	}
	unsignedVal, err := p.connection.ReadByte()
	if err != nil {
		return 0, err
	}

	value = int(unsignedVal)
	if mc.isDiff() {
		value = value - 128
	}

	return value, err
}

// AnalogWrite writes the given value to the analog output (DAC)
// Vlsb = (Vref-Vagnd)/256
// Vaout = Vagnd+Vlsb*value
func (p *PCF8591Driver) AnalogWrite(value uint8) (err error) {
	if err = p.AnalogOutputState(true); err != nil {
		return err
	}
	if err = p.connection.WriteByte(value); err != nil {
		return err
	}
	return nil
}

// AnalogOutputState enables or disables the analog output
// Please note that in case of using the internal oscillator
// and the auto increment mode the output should not switched off.
// Otherwise conversion errors could occur.
func (p *PCF8591Driver) AnalogOutputState(state bool) (err error) {
	var ctrlByte uint8
	if state {
		ctrlByte = uint8(pcf8591_DAMASK | pcf8591_ANAON)
	} else {
		ctrlByte = uint8(pcf8591_DAMASK & ^pcf8591_ANAON)
	}
	if err = p.connection.WriteByte(ctrlByte); err != nil {
		return err
	}
	return nil
}

func (mc modeChan) isDiff() bool {
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
