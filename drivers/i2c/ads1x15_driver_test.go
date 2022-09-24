package i2c

import (
	"errors"
	"strings"
	"testing"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/gobottest"
)

// the ADS1x15Driver is a Driver
var _ gobot.Driver = (*ADS1x15Driver)(nil)

// that supports the AnalogReader interface
var _ aio.AnalogReader = (*ADS1x15Driver)(nil)

// --------- HELPERS
func initTestADS1015Driver() (driver *ADS1x15Driver) {
	driver, _ = initTestADS1015DriverWithStubbedAdaptor()
	return
}

func initTestADS1115Driver() (driver *ADS1x15Driver) {
	driver, _ = initTestADS1115DriverWithStubbedAdaptor()
	return
}

func initTestADS1015DriverWithStubbedAdaptor() (*ADS1x15Driver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor()
	return NewADS1015Driver(adaptor), adaptor
}

func initTestADS1115DriverWithStubbedAdaptor() (*ADS1x15Driver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor()
	return NewADS1115Driver(adaptor), adaptor
}

// --------- BASE TESTS
func TestNewADS1015Driver(t *testing.T) {
	// Does it return a pointer to an instance of ADS1x15Driver?
	var bm interface{} = NewADS1015Driver(newI2cTestAdaptor())
	_, ok := bm.(*ADS1x15Driver)
	if !ok {
		t.Errorf("NewADS1015Driver() should have returned a *ADS1x15Driver")
	}
}

func TestNewADS1115Driver(t *testing.T) {
	// Does it return a pointer to an instance of ADS1x15Driver?
	var bm interface{} = NewADS1115Driver(newI2cTestAdaptor())
	_, ok := bm.(*ADS1x15Driver)
	if !ok {
		t.Errorf("NewADS1115Driver() should have returned a *ADS1x15Driver")
	}
}

func TestADS1x15DriverSetName(t *testing.T) {
	d := initTestADS1015Driver()
	d.SetName("TESTME")
	gobottest.Assert(t, d.Name(), "TESTME")
}

func TestADS1x15DriverOptions(t *testing.T) {
	d := NewADS1015Driver(newI2cTestAdaptor(), WithBus(2), WithADS1x15Gain(2), WithADS1x15DataRate(920))
	gobottest.Assert(t, d.GetBusOrDefault(1), 2)
	gobottest.Assert(t, d.DefaultGain, 2)
	gobottest.Assert(t, d.DefaultDataRate, 920)
}

func TestADS1x15StartAndHalt(t *testing.T) {
	d, _ := initTestADS1015DriverWithStubbedAdaptor()
	gobottest.Assert(t, d.Start(), nil)
	gobottest.Refute(t, d.Connection(), nil)
	gobottest.Assert(t, d.Halt(), nil)
}

func TestADS1x15StartConnectError(t *testing.T) {
	d, adaptor := initTestADS1015DriverWithStubbedAdaptor()
	adaptor.Testi2cConnectErr(true)
	gobottest.Assert(t, d.Start(), errors.New("Invalid i2c connection"))
}

// --------- DRIVER SPECIFIC TESTS

func TestADS1015DriverAnalogRead(t *testing.T) {
	d, adaptor := initTestADS1015DriverWithStubbedAdaptor()
	WithADS1x15WaitSingleCycle()(d)
	d.Start()

	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		copy(b, []byte{0x7F, 0xFF})
		return 2, nil
	}

	val, err := d.AnalogRead("0")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("1")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("2")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("3")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("0-1")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("0-3")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("1-3")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("2-3")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("3-2")
	gobottest.Refute(t, err.Error(), nil)
}

func TestADS1115DriverAnalogRead(t *testing.T) {
	d, adaptor := initTestADS1115DriverWithStubbedAdaptor()
	WithADS1x15WaitSingleCycle()(d)
	d.Start()

	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		copy(b, []byte{0x7F, 0xFF})
		return 2, nil
	}

	val, err := d.AnalogRead("0")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("1")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("2")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("3")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("0-1")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("0-3")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("1-3")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("2-3")
	gobottest.Assert(t, val, 1022)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("3-2")
	gobottest.Refute(t, err.Error(), nil)
}

func TestADS1x15DriverAnalogReadError(t *testing.T) {
	d, a := initTestADS1015DriverWithStubbedAdaptor()
	d.Start()

	a.i2cReadImpl = func(b []byte) (int, error) {
		return 0, errors.New("read error")
	}

	_, err := d.AnalogRead("0")
	gobottest.Assert(t, err, errors.New("read error"))
}

func TestADS1x15DriverAnalogReadInvalidPin(t *testing.T) {
	d, _ := initTestADS1015DriverWithStubbedAdaptor()

	_, err := d.AnalogRead("99")
	gobottest.Assert(t, err, errors.New("Invalid channel, must be between 0 and 3"))
}

func TestADS1x15DriverAnalogReadWriteError(t *testing.T) {
	d, a := initTestADS1015DriverWithStubbedAdaptor()
	d.Start()

	a.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}

	_, err := d.AnalogRead("0")
	gobottest.Assert(t, err, errors.New("write error"))

	_, err = d.AnalogRead("0-1")
	gobottest.Assert(t, err, errors.New("write error"))

	_, err = d.AnalogRead("2-3")
	gobottest.Assert(t, err, errors.New("write error"))
}

func TestADS1x15DriverBestGainForVoltage(t *testing.T) {
	d, _ := initTestADS1015DriverWithStubbedAdaptor()

	g, err := d.BestGainForVoltage(1.5)
	gobottest.Assert(t, g, 2)

	g, err = d.BestGainForVoltage(20.0)
	gobottest.Assert(t, err, errors.New("The maximum voltage which can be read is 6.144000"))
}

func TestADS1x15DriverReadInvalidChannel(t *testing.T) {
	d, _ := initTestADS1015DriverWithStubbedAdaptor()

	_, err := d.Read(9, d.DefaultGain, d.DefaultDataRate)
	gobottest.Assert(t, err, errors.New("Invalid channel, must be between 0 and 3"))
}

func TestADS1x15DriverReadInvalidGain(t *testing.T) {
	d, _ := initTestADS1015DriverWithStubbedAdaptor()

	_, err := d.Read(0, 21, d.DefaultDataRate)
	gobottest.Assert(t, err, errors.New("Gain must be one of: 2/3, 1, 2, 4, 8, 16"))
}

func TestADS1x15DriverReadInvalidDataRate(t *testing.T) {
	d, _ := initTestADS1015DriverWithStubbedAdaptor()

	_, err := d.Read(0, d.DefaultGain, 666)
	gobottest.Assert(t, strings.Contains(err.Error(), "Invalid data rate."), true)
}

func TestADS1x15DriverReadDifferenceInvalidChannel(t *testing.T) {
	d, _ := initTestADS1015DriverWithStubbedAdaptor()

	_, err := d.ReadDifference(9, d.DefaultGain, d.DefaultDataRate)
	gobottest.Assert(t, err, errors.New("Invalid channel, must be between 0 and 3"))
}

func TestADS1x15Driver_rawRead(t *testing.T) {
	// sequence to read:
	// * prepare config register content (mode, input, gain, data rate, comparator)
	// * write config register (16 bit, MSByte first)
	// * read config register (16 bit, MSByte first) and wait for bit 15 is set
	// * read conversion register (16 bit, MSByte first) for the value
	// * apply two's complement converter, relates to one digit resolution (1/2^15), voltage multiplier
	var tests = map[string]struct {
		input      []uint8
		gain       int
		dataRate   int
		want       int
		wantConfig []uint8
	}{
		"+FS": {
			input:      []uint8{0x7F, 0xFF},
			gain:       0,
			dataRate:   8,
			want:       (1<<15 - 1),
			wantConfig: []uint8{0x91, 0x03},
		},
		"+1": {
			input:      []uint8{0x00, 0x01},
			gain:       0,
			dataRate:   16,
			want:       1,
			wantConfig: []uint8{0x91, 0x23},
		},
		"+-0": {
			input:      []uint8{0x00, 0x00},
			gain:       0,
			dataRate:   32,
			want:       0,
			wantConfig: []uint8{0x91, 0x43},
		},
		"-1": {
			input:      []uint8{0xFF, 0xFF},
			gain:       0,
			dataRate:   64,
			want:       -1,
			wantConfig: []uint8{0x91, 0x63},
		},
		"-FS": {
			input:      []uint8{0x80, 0x00},
			gain:       0,
			dataRate:   128,
			want:       -(1 << 15),
			wantConfig: []uint8{0x91, 0x83},
		},
		"+FS gain 1": {
			input:      []uint8{0x7F, 0xFF},
			gain:       1,
			dataRate:   250,
			want:       (1<<15 - 1),
			wantConfig: []uint8{0x93, 0xA3},
		},
		"+FS gain 2": {
			input:      []uint8{0x7F, 0xFF},
			gain:       2,
			dataRate:   250,
			want:       (1<<15 - 1),
			wantConfig: []uint8{0x95, 0xA3},
		},
		"+FS gain 4": {
			input:      []uint8{0x7F, 0xFF},
			gain:       4,
			dataRate:   475,
			want:       (1<<15 - 1),
			wantConfig: []uint8{0x97, 0xC3},
		},
		"+FS gain 8": {
			input:      []uint8{0x7F, 0xFF},
			gain:       8,
			dataRate:   860,
			want:       (1<<15 - 1),
			wantConfig: []uint8{0x99, 0xE3},
		},
		"+FS gain 16": {
			input:      []uint8{0x7F, 0xFF},
			gain:       16,
			dataRate:   128,
			want:       (1<<15 - 1),
			wantConfig: []uint8{0x9B, 0x83},
		},
	}
	d, adaptor := initTestADS1115DriverWithStubbedAdaptor()
	d.Start()
	// arrange
	channel := 1
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			adaptor.written = []byte{} // reset writes of Start() and former test
			// arrange reads
			conversion := []uint8{0x00, 0x00}   // a conversion is in progress
			noConversion := []uint8{0x80, 0x00} // no conversion in progress
			returnRead := [3][]uint8{conversion, noConversion, tt.input}
			numCallsRead := 0
			adaptor.i2cReadImpl = func(b []byte) (int, error) {
				numCallsRead++
				retRead := returnRead[numCallsRead-1]
				copy(b, retRead)
				return len(b), nil
			}
			// act
			got, err := d.rawRead(channel, tt.gain, tt.dataRate)
			// assert
			gobottest.Assert(t, err, nil)
			gobottest.Assert(t, got, float64(tt.want)/float64(1<<15)*d.gainVoltage[tt.gain])
			gobottest.Assert(t, numCallsRead, 3)
			gobottest.Assert(t, len(adaptor.written), 6)
			gobottest.Assert(t, adaptor.written[0], uint8(ads1x15PointerConfig))
			gobottest.Assert(t, adaptor.written[1], tt.wantConfig[0])            // MSByte: OS, MUX, PGA, MODE
			gobottest.Assert(t, adaptor.written[2], tt.wantConfig[1])            // LSByte: DR, COMP_*
			gobottest.Assert(t, adaptor.written[3], uint8(ads1x15PointerConfig)) // first check for no conversion
			gobottest.Assert(t, adaptor.written[4], uint8(ads1x15PointerConfig)) // second check for no conversion
			gobottest.Assert(t, adaptor.written[5], uint8(ads1x15PointerConversion))
		})
	}
}
