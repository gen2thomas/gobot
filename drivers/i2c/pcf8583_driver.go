package i2c

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gobot.io/x/gobot"
)

// PCF8583 supports addresses 0x50 and 0x51
// The default address applies when the address pin is grounded.
const pcf8583Address = 0x50

const pcf8583Debug = true

type PCF8583Reg uint8

const (
	pcf8583_CTRL         PCF8583Reg = iota // 0x00
	pcf8583_SUBSEC_D0D1                    // 0x01
	pcf8583_SEC_D2D3                       // 0x02
	pcf8583_MIN_D4D5                       // 0x03
	pcf8583_HOUR                           // 0x04
	pcf8583_YEARDATE                       // 0x05
	pcf8583_WEEKDAYMONTH                   // 0x06
	pcf8583_TIMER                          // 0x07
	pcf8583_ALARMCTRL                      // 0x08, offset for all alarm registers 0x09 ... 0xF
)

// PCF8583Control is used to specify control and status register content
type PCF8583Control uint8

const (
	pcf8583TimerFlag     PCF8583Control = 0x01 // 50% duty factor, seconds flag if alarm enable bit is 0
	pcf8583AlarmFlag     PCF8583Control = 0x02 // 50% duty factor, minutes flag if alarm enable bit is 0
	pcf8583AlarmEnable   PCF8583Control = 0x04 // if enabled, memory 08h is alarm control register
	pcf8583Mask          PCF8583Control = 0x08 // 0: read 05h, 06h unmasked, 1: read date and month count directly
	PCF8583ModeClock50   PCF8583Control = 0x10 // clock mode with 50 Hz
	PCF8583ModeCounter   PCF8583Control = 0x20 // event counter mode
	PCF8583ModeTest      PCF8583Control = 0x30 // test mode
	pcf8583HoldLastCount PCF8583Control = 0x40 // 0: count, 1: store and hold count in capture latches
	pcf8583StopCounting  PCF8583Control = 0x80 // 0: count, 1: stop counting, reset divider
)

// default is 0x10, when set to 0 also some free or unused ram can be accessed
const pcf8583RamOffset = 0x10

// PCF8583Driver is a Gobot Driver for the PCF8583 clock and calendar chip & 240 x 8-bit bit RAM with 1 address program pin.
// please refer to data sheet: https://www.nxp.com/docs/en/data-sheet/PCF8583.pdf
//
// 0 1 0 1 0 0 0 A0|rd
// Lowest bit (rd) is mapped to switch between write(0)/read(1), it is not part of the "real" address.
//
// PCF8583 is mainly compatible to PCF8593, so this driver should also work for PCF8593 except RAM calls
//
type PCF8583Driver struct {
	name       string
	mode       PCF8583Control // clock 32.768kHz (default), clock 50Hz, event counter
	yearOffset int
	ramOffset  byte
	connector  Connector
	connection Connection
	Config
	gobot.Commander
	mutex *sync.Mutex // mutex needed to ensure write-read sequences are not interrupted
}

// NewPCF8583Driver creates a new driver with specified i2c interface
// Params:
//		conn Connector - the Adaptor to use with this Driver
//
// Optional params:
//		i2c.WithBus(int):	bus to use with this driver
//		i2c.WithAddress(int):	address to use with this driver
//    i2c.WithPCF8583Mode(PCF8583Control): mode of this driver
//
func NewPCF8583Driver(a Connector, options ...func(Config)) *PCF8583Driver {
	p := &PCF8583Driver{
		name:      gobot.DefaultName("PCF8583"),
		connector: a,
		Config:    NewConfig(),
		Commander: gobot.NewCommander(),
		ramOffset: pcf8583RamOffset,
		mutex:     &sync.Mutex{},
	}

	for _, option := range options {
		option(p)
	}

	// API commands
	p.AddCommand("WriteTime", func(params map[string]interface{}) interface{} {
		val := params["val"].(time.Time)
		err := p.WriteTime(val)
		return map[string]interface{}{"err": err}
	})

	p.AddCommand("ReadTime", func(params map[string]interface{}) interface{} {
		val, err := p.ReadTime()
		return map[string]interface{}{"val": val, "err": err}
	})

	p.AddCommand("WriteCounter", func(params map[string]interface{}) interface{} {
		val := params["val"].(int32)
		err := p.WriteCounter(val)
		return map[string]interface{}{"err": err}
	})

	p.AddCommand("ReadCounter", func(params map[string]interface{}) interface{} {
		val, err := p.ReadCounter()
		return map[string]interface{}{"val": val, "err": err}
	})

	p.AddCommand("WriteRAM", func(params map[string]interface{}) interface{} {
		address := params["address"].(uint8)
		val := params["val"].(uint8)
		err := p.WriteRAM(address, val)
		return map[string]interface{}{"err": err}
	})

	p.AddCommand("ReadRAM", func(params map[string]interface{}) interface{} {
		address := params["address"].(uint8)
		val, err := p.ReadRAM(address)
		return map[string]interface{}{"val": val, "err": err}
	})
	return p
}

// WithPCF8583Mode is used to change the mode between 32.678kHz clock, 50Hz clock, event counter
func WithPCF8583Mode(mode PCF8583Control) func(Config) {
	return func(c Config) {
		p, ok := c.(*PCF8583Driver)
		if ok {
			if !mode.isClockMode() && !mode.isCounterMode() {
				panic(fmt.Sprintf("%s: mode 0x%02x is not supported", p.name, mode))
			}
			p.mode = mode
		} else if pcf8583Debug {
			log.Printf("trying to set mode for non-PCF8583Driver %v", c)
		}
	}
}

// Name returns the Name for the Driver
func (p *PCF8583Driver) Name() string { return p.name }

// SetName sets the Name for the Driver
func (p *PCF8583Driver) SetName(n string) { p.name = n }

// Connection returns the connection for the Driver
func (p *PCF8583Driver) Connection() gobot.Connection { return p.connector.(gobot.Connection) }

// Start initializes the driver
func (p *PCF8583Driver) Start() (err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	bus := p.GetBusOrDefault(p.connector.GetDefaultBus())
	address := p.GetAddressOrDefault(pcf8583Address)
	if p.connection, err = p.connector.GetConnection(address, bus); err != nil {
		return
	}

	// switch to configured mode
	ctrlRegVal, err := p.connection.ReadByteData(uint8(pcf8583_CTRL))
	if err != nil {
		return
	}
	if p.mode.isModeDiffer(PCF8583Control(ctrlRegVal)) {
		ctrlRegVal = ctrlRegVal&^uint8(PCF8583ModeTest) | uint8(p.mode)
		if err = p.connection.WriteByteData(uint8(pcf8583_CTRL), ctrlRegVal); err != nil {
			return
		}
		if pcf8583Debug {
			if PCF8583Control(ctrlRegVal).isCounterMode() {
				log.Printf("%s switched to counter mode 0x%02x", p.name, ctrlRegVal)
			} else {
				log.Printf("%s switched to clock mode 0x%02x", p.name, ctrlRegVal)
			}
		}
	}
	return
}

// Halt stops the device
func (p *PCF8583Driver) Halt() (err error) { return }

// WriteTime setup the clock registers with the given time
func (p *PCF8583Driver) WriteTime(val time.Time) (err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// according to chapter 7.11 of the product data sheet, the stop counting flag of the control/status register
	// must be set before, so we read the control byte before and only set/reset the stop
	ctrlRegVal, err := p.connection.ReadByteData(uint8(pcf8583_CTRL))
	if err != nil {
		return
	}
	if !PCF8583Control(ctrlRegVal).isClockMode() {
		return fmt.Errorf("%s: can't write time because the device is in wrong mode 0x%02x", p.name, ctrlRegVal)
	}
	// auto increment feature is used
	year, month, day := val.Date()
	written, err := p.connection.Write([]byte{
		uint8(pcf8583_CTRL), ctrlRegVal | uint8(pcf8583StopCounting),
		pcf8583encodeBcd(uint8(val.Nanosecond() / 1000000 / 10)), // sub seconds in 1/10th seconds
		pcf8583encodeBcd(uint8(val.Second())),
		pcf8583encodeBcd(uint8(val.Minute())),
		pcf8583encodeBcd(uint8(val.Hour())),
		pcf8583encodeBcd(uint8(day)),                             // year, date (we keep the year counter zero and set the offset)
		uint8(val.Weekday())<<5 | pcf8583encodeBcd(uint8(month)), // month, weekday (not BCD): Sunday = 0, Monday = 1 ...
	})
	if err != nil {
		return
	}
	if written != 8 {
		return fmt.Errorf("%s: %d bytes written, but %d expected", p.name, written, 8)
	}
	p.yearOffset = year
	return p.run(ctrlRegVal)
}

// ReadTime reads the clock and returns the value
func (p *PCF8583Driver) ReadTime() (val time.Time, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// according to chapter 7.1 of the product data sheet, the setting of "hold last count" flag
	// is not needed when reading with auto increment
	ctrlRegVal, err := p.connection.ReadByteData(uint8(pcf8583_CTRL))
	if err != nil {
		return
	}
	if !PCF8583Control(ctrlRegVal).isClockMode() {
		return val, fmt.Errorf("%s: can't read time because the device is in wrong mode 0x%02x", p.name, ctrlRegVal)
	}
	// auto increment feature is used
	clockDataSize := 6
	data := make([]byte, clockDataSize)
	read, err := p.connection.Read(data)
	if err != nil {
		return
	}
	if read != clockDataSize {
		return val, fmt.Errorf("%s: %d bytes read, but %d expected", p.name, read, clockDataSize)
	}
	nanos := int(pcf8583decodeBcd(data[0])) * 1000000 * 10 // sub seconds in 1/10th seconds
	seconds := int(pcf8583decodeBcd(data[1]))
	minutes := int(pcf8583decodeBcd(data[2]))
	hours := int(pcf8583decodeBcd(data[3]))
	// year, date (the device can only count 4 years)
	year := int(data[4]>>6) + p.yearOffset        // use the first two bits, no BCD
	date := int(pcf8583decodeBcd(data[4] & 0x3F)) // remove the year-bits for date
	// weekday (not used here), month
	month := time.Month(pcf8583decodeBcd(data[5] & 0x1F)) // remove the weekday-bits
	return time.Date(year, month, date, hours, minutes, seconds, nanos, time.UTC), nil
}

// WriteCounter writes the counter registers
func (p *PCF8583Driver) WriteCounter(val int32) (err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// we don't care of negative values here
	// according to chapter 7.11 of the product data sheet, the stop counting flag of the control/status register
	// must be set before, so we read the control byte before and only set/reset the stop
	ctrlRegVal, err := p.connection.ReadByteData(uint8(pcf8583_CTRL))
	if err != nil {
		return
	}
	if !PCF8583Control(ctrlRegVal).isCounterMode() {
		return fmt.Errorf("%s: can't write counter because the device is in wrong mode 0x%02x", p.name, ctrlRegVal)
	}
	// auto increment feature is used
	written, err := p.connection.Write([]byte{
		uint8(pcf8583_CTRL), ctrlRegVal | uint8(pcf8583StopCounting), // stop
		pcf8583encodeBcd(uint8(val % 100)),           // 2 lowest digits
		pcf8583encodeBcd(uint8((val / 100) % 100)),   // 2 middle digits
		pcf8583encodeBcd(uint8((val / 10000) % 100)), // 2 highest digits
	})
	if err != nil {
		return
	}
	if written != 5 {
		return fmt.Errorf("%s: %d bytes written, but %d expected", p.name, written, 5)
	}
	return p.run(ctrlRegVal)
}

// ReadCounter reads the counter registers
func (p *PCF8583Driver) ReadCounter() (val int32, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// according to chapter 7.1 of the product data sheet, the setting of "hold last count" flag
	// is not needed when reading with auto increment
	ctrlRegVal, err := p.connection.ReadByteData(uint8(pcf8583_CTRL))
	if err != nil {
		return
	}
	if !PCF8583Control(ctrlRegVal).isCounterMode() {
		return val, fmt.Errorf("%s: can't read counter because the device is in wrong mode 0x%02x", p.name, ctrlRegVal)
	}
	// auto increment feature is used
	counterDataSize := 3
	data := make([]byte, counterDataSize)
	read, err := p.connection.Read(data)
	if err != nil {
		return
	}
	if read != counterDataSize {
		return val, fmt.Errorf("%s: %d bytes read, but %d expected", p.name, read, counterDataSize)
	}
	return int32(pcf8583decodeBcd(data[0])) +
		int32(pcf8583decodeBcd(data[1]))*100 +
		int32(pcf8583decodeBcd(data[2]))*10000, nil
}

// WriteRAM writes a value to a given address in memory (0x00-0xFF)
func (p *PCF8583Driver) WriteRAM(address uint8, val uint8) (err error) {
	realAddress := uint16(address) + uint16(p.ramOffset)
	if realAddress > 0xFF {
		return fmt.Errorf("%s: RAM address overflow %d", p.name, realAddress)
	}
	return p.connection.WriteByteData(uint8(realAddress), val)
}

// ReadRAM reads a value from a given address (0x00-0xFF)
func (p *PCF8583Driver) ReadRAM(address uint8) (val uint8, err error) {
	realAddress := uint16(address) + uint16(p.ramOffset)
	if realAddress > 0xFF {
		return val, fmt.Errorf("%s: RAM address overflow %d", p.name, realAddress)
	}
	return p.connection.ReadByteData(uint8(realAddress))
}

func (p *PCF8583Driver) run(ctrlRegVal uint8) error {
	ctrlRegVal = ctrlRegVal & ^uint8(pcf8583StopCounting) // reset stop bit
	return p.connection.WriteByteData(uint8(pcf8583_CTRL), ctrlRegVal)
}

func (c PCF8583Control) isClockMode() bool {
	return uint8(c)&uint8(PCF8583ModeCounter) == 0
}

func (c PCF8583Control) isCounterMode() bool {
	counterModeSet := uint8(c) & uint8(PCF8583ModeCounter)
	clockMode50Set := uint8(c) & uint8(PCF8583ModeClock50)
	return counterModeSet > 0 && clockMode50Set == 0
}

func (c PCF8583Control) isModeDiffer(mode PCF8583Control) bool {
	return uint8(c)&uint8(PCF8583ModeTest) != uint8(mode)&uint8(PCF8583ModeTest)
}

func pcf8583encodeBcd(val byte) byte {
	// decimal 12 => 0x12
	if val > 99 {
		val = 99
		if pcf8583Debug {
			log.Printf("PCF8583 BCD value (%d) exceeds limit of 99, now limited.", val)
		}
	}
	hi, lo := byte(val/10), byte(val%10)
	return hi<<4 | lo
}

func pcf8583decodeBcd(bcd byte) byte {
	// 0x12 => decimal 12
	hi, lo := byte(bcd>>4), byte(bcd&0x0f)
	if hi > 9 {
		hi = 9
		if pcf8583Debug {
			log.Printf("PCF8583 BCD value (%02x) exceeds limit 0x99 on most significant digit, now limited", bcd)
		}
	}
	if lo > 9 {
		lo = 9
		if pcf8583Debug {
			log.Printf("PCF8583 BCD value (%02x) exceeds limit 0x99 on least significant digit, now limited", bcd)
		}
	}
	return 10*hi + lo
}
