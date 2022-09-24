package i2c

import (
	"errors"
	"testing"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/gobottest"
)

var _ gobot.Driver = (*HMC5883LDriver)(nil)

// --------- HELPERS
func initTestHMC5883LDriver() (driver *HMC5883LDriver) {
	driver, _ = initTestHMC5883LDriverWithStubbedAdaptor()
	return
}

func initTestHMC5883LDriverWithStubbedAdaptor() (*HMC5883LDriver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor()
	return NewHMC5883LDriver(adaptor), adaptor
}

// --------- TESTS

func TestNewHMC5883LDriver(t *testing.T) {
	// Does it return a pointer to an instance of HMC5883LDriver?
	var bm interface{} = NewHMC5883LDriver(newI2cTestAdaptor())
	_, ok := bm.(*HMC5883LDriver)
	if !ok {
		t.Errorf("NewHMC5883LDriver() should have returned a *HMC5883LDriver")
	}

	d := NewHMC5883LDriver(newI2cTestAdaptor())
	gobottest.Refute(t, d.Connection(), nil)
}

func TestHMC5883LStart(t *testing.T) {
	// sequence to prepare read in Start():
	// * prepare config register A content (samples averaged, data output rate, measurement mode)
	// * prepare config register B content (gain)
	// * prepare mode register (continuous/single/idle)
	// * write registers A, B, mode
	// arrange
	d, adaptor := initTestHMC5883LDriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of former test
	wantRegA := uint8(0x70)
	wantRegB := uint8(0xA0)
	wantRegM := uint8(0x00)
	// act
	err := d.Start()
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, len(adaptor.written), 6)
	gobottest.Assert(t, adaptor.written[0], uint8(hmc5883lRegA))
	gobottest.Assert(t, adaptor.written[1], wantRegA)
	gobottest.Assert(t, adaptor.written[2], uint8(hmc5883lRegB))
	gobottest.Assert(t, adaptor.written[3], wantRegB)
	gobottest.Assert(t, adaptor.written[4], uint8(hmc5883lRegMode))
	gobottest.Assert(t, adaptor.written[5], wantRegM)
}

func TestHMC5883LStartError(t *testing.T) {
	d, adaptor := initTestHMC5883LDriverWithStubbedAdaptor()

	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}
	err := d.Start()
	gobottest.Assert(t, err, errors.New("write error"))
	gobottest.Assert(t, d.measurementMode, hmc5883lRegM_Continuous)
}

func TestHMC5883LStartConnectError(t *testing.T) {
	d, adaptor := initTestHMC5883LDriverWithStubbedAdaptor()
	adaptor.Testi2cConnectErr(true)
	gobottest.Assert(t, d.Start(), errors.New("Invalid i2c connection"))
}

func TestHMC5883LHalt(t *testing.T) {
	d := initTestHMC5883LDriver()
	gobottest.Assert(t, d.Halt(), nil)
}

func TestHMC5883LSetName(t *testing.T) {
	d := initTestHMC5883LDriver()
	d.SetName("TESTME")
	gobottest.Assert(t, d.Name(), "TESTME")
}

func TestHMC5883LWithBus(t *testing.T) {
	d := NewHMC5883LDriver(newI2cTestAdaptor())
	gobottest.Assert(t, d.GetBusOrDefault(1), 1)

	WithBus(2)(d)
	gobottest.Assert(t, d.GetBusOrDefault(1), 2)
}

func TestHMC5883LWithHMC5883LSamplesAveraged(t *testing.T) {
	d := NewHMC5883LDriver(newI2cTestAdaptor())
	gobottest.Assert(t, d.samplesAvg, uint8(8))

	WithHMC5883LSamplesAveraged(4)(d)
	gobottest.Assert(t, d.samplesAvg, uint8(4))
}

func TestHMC5883LWithHMC5883LDataOutputRate(t *testing.T) {
	d := NewHMC5883LDriver(newI2cTestAdaptor())
	gobottest.Assert(t, d.outputRate, uint32(15000))

	WithHMC5883LDataOutputRate(7500)(d)
	gobottest.Assert(t, d.outputRate, uint32(7500))
}

func TestHMC5883LWithHMC5883LApplyBias(t *testing.T) {
	d := NewHMC5883LDriver(newI2cTestAdaptor())
	gobottest.Assert(t, d.applyBias, int8(0))

	WithHMC5883LApplyBias(-1)(d)
	gobottest.Assert(t, d.applyBias, int8(-1))
}

func TestHMC5883LWithHMC5883LGain(t *testing.T) {
	d := NewHMC5883LDriver(newI2cTestAdaptor())
	gobottest.Assert(t, d.gain, 390.0)

	WithHMC5883LGain(230)(d)
	gobottest.Assert(t, d.gain, 230.0)
}

func TestHMC5883LReadRawData(t *testing.T) {
	// sequence to read:
	// * prepare read, see test of Start()
	// * read data output registers (3 x 16 bit, MSByte first)
	// * apply two's complement converter
	//
	// arrange
	var tests = map[string]struct {
		inputX []uint8
		inputY []uint8
		inputZ []uint8
		wantX  int16
		wantY  int16
		wantZ  int16
	}{
		"+FS_0_-FS": {
			inputX: []uint8{0x07, 0xFF},
			inputY: []uint8{0x00, 0x00},
			inputZ: []uint8{0xF8, 0x00},
			wantX:  (1<<11 - 1),
			wantY:  0,
			wantZ:  -(1 << 11),
		},
		"-4096_-1_+1": {
			inputX: []uint8{0xF0, 0x00},
			inputY: []uint8{0xFF, 0xFF},
			inputZ: []uint8{0x00, 0x01},
			wantX:  -4096,
			wantY:  -1,
			wantZ:  1,
		},
	}
	d, adaptor := initTestHMC5883LDriverWithStubbedAdaptor()
	d.Start()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			adaptor.written = []byte{} // reset writes of former test and start
			// arrange reads
			returnRead := append(append(tc.inputX, tc.inputZ...), tc.inputY...)
			numCallsRead := 0
			adaptor.i2cReadImpl = func(b []byte) (int, error) {
				numCallsRead++
				copy(b, returnRead)
				return len(b), nil
			}
			// act
			gotX, gotY, gotZ, err := d.ReadRawData()
			// assert
			gobottest.Assert(t, err, nil)
			gobottest.Assert(t, gotX, tc.wantX)
			gobottest.Assert(t, gotY, tc.wantY)
			gobottest.Assert(t, gotZ, tc.wantZ)
			gobottest.Assert(t, numCallsRead, 1)
			gobottest.Assert(t, len(adaptor.written), 1)
			gobottest.Assert(t, adaptor.written[0], uint8(hmc5883lAxisX))
		})
	}
}

func TestHMC5883LRead(t *testing.T) {
	// arrange
	var tests = map[string]struct {
		inputX []uint8
		inputY []uint8
		inputZ []uint8
		gain   float64
		wantX  float64
		wantY  float64
		wantZ  float64
	}{
		"+FS_0_-FS_resolution_0.73mG": {
			inputX: []uint8{0x07, 0xFF},
			inputY: []uint8{0x00, 0x00},
			inputZ: []uint8{0xF8, 0x00},
			gain:   1370,
			wantX:  2047.0 / 1370,
			wantY:  0,
			wantZ:  -2048.0 / 1370,
		},
		"+1_-4096_-1_resolution_0.73mG": {
			inputX: []uint8{0x00, 0x01},
			inputY: []uint8{0xF0, 0x00},
			inputZ: []uint8{0xFF, 0xFF},
			gain:   1370,
			wantX:  1.0 / 1370,
			wantY:  -4096.0 / 1370,
			wantZ:  -1.0 / 1370,
		},
		"+FS_0_-FS_resolution_4.35mG": {
			inputX: []uint8{0x07, 0xFF},
			inputY: []uint8{0x00, 0x00},
			inputZ: []uint8{0xF8, 0x00},
			gain:   230,
			wantX:  2047.0 / 230,
			wantY:  0,
			wantZ:  -2048.0 / 230,
		},
		"-1_+1_-4096_resolution_4.35mG": {
			inputX: []uint8{0xFF, 0xFF},
			inputY: []uint8{0x00, 0x01},
			inputZ: []uint8{0xF0, 0x00},
			gain:   230,
			wantX:  -1.0 / 230,
			wantY:  1.0 / 230,
			wantZ:  -4096.0 / 230,
		},
	}
	adaptor := newI2cTestAdaptor()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d := NewHMC5883LDriver(adaptor, WithHMC5883LGain(int(tc.gain)))
			d.Start()
			// arrange reads
			returnRead := append(append(tc.inputX, tc.inputZ...), tc.inputY...)
			adaptor.i2cReadImpl = func(b []byte) (int, error) {
				copy(b, returnRead)
				return len(b), nil
			}
			// act
			gotX, gotY, gotZ, err := d.Read()
			// assert
			gobottest.Assert(t, err, nil)
			gobottest.Assert(t, gotX, tc.wantX)
			gobottest.Assert(t, gotY, tc.wantY)
			gobottest.Assert(t, gotZ, tc.wantZ)
		})
	}
}
