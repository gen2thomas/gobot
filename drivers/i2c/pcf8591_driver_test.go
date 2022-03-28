package i2c

import (
	"testing"

	"gobot.io/x/gobot/gobottest"
)

func initTestPCF8591DriverWithStubbedAdaptor() (*PCF8591Driver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor()
	pcf := NewPCF8591Driver(adaptor)
	pcf.lastCtrlByte = 0xFF // prevent skipping of write
	pcf.Start()
	return pcf, adaptor
}

func TestPCF8591DriverWithAdditionalSkip(t *testing.T) {
	pcf := NewPCF8591Driver(newI2cTestAdaptor(), WithPCF8591AdditionalSkip(5))
	gobottest.Assert(t, pcf.additionalSkip, uint8(5))
}

func TestPCF8591DriverWithRescaleInput(t *testing.T) {
	pcf := NewPCF8591Driver(newI2cTestAdaptor(), WithPCF8591RescaleInput(2, -3, 4))
	gobottest.Assert(t, pcf.toMin, [4]int{0, 0, -3, 0})
	gobottest.Assert(t, pcf.toMax, [4]int{3300, 3300, 4, 3300})
}

func TestPCF8591DriverWithRescaleOutput(t *testing.T) {
	pcf := NewPCF8591Driver(newI2cTestAdaptor(), WithPCF8591RescaleOutput(5, 178))
	gobottest.Assert(t, pcf.fromMin, 5)
	gobottest.Assert(t, pcf.fromMax, 178)
}

func TestPCF8591DriverAnalogReadSingle(t *testing.T) {
	// sequence to read the input channel:
	// * prepare value (with channel and mode) and write control register
	// * read 3 values to drop (see description in implementation)
	// * read the analog value
	//
	// arrange
	pcf, adaptor := initTestPCF8591DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	description := "s.1"
	pcf.toMin[1] = 0
	pcf.toMax[1] = 255
	pcf.lastCtrlByte = 0x00
	ctrlByteOn := uint8(pcf8591_ALLSINGLE | pcf8591_CHAN1)
	returnRead := []uint8{0x01, 0x02, 0x03, 0xFF}
	want := int(returnRead[3])
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		if numCallsRead == 1 {
			b = returnRead[0:len(b)]
		}
		if numCallsRead == 2 {
			b[0] = returnRead[len(returnRead)-1]
		}
		return len(b), nil
	}
	// act
	got, err := pcf.AnalogRead(description)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], ctrlByteOn)
	gobottest.Assert(t, numCallsRead, 2)
	gobottest.Assert(t, got, want)
}

func TestPCF8591DriverAnalogReadDiff(t *testing.T) {
	// sequence to read the input channel:
	// * prepare value (with channel and mode) and write control register
	// * read 3 values to drop (see description in implementation)
	// * read the analog value
	// * convert to 8-bit two's complement (-127...128)
	//
	// arrange
	pcf, adaptor := initTestPCF8591DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	description := "m.2-3"
	pcf.toMin[2] = -128
	pcf.toMax[2] = 127
	pcf.lastCtrlByte = 0x00
	ctrlByteOn := uint8(pcf8591_MIXED | pcf8591_CHAN2)
	// some two' complements
	// 0x80 => -128
	// 0xFF => -1
	// 0x00 => 0
	// 0x7F => 127
	returnRead := []uint8{0x01, 0x02, 0x03, 0xFF}
	want := -1
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		if numCallsRead == 1 {
			b = returnRead[0:len(b)]
		}
		if numCallsRead == 2 {
			b[0] = returnRead[len(returnRead)-1]
		}
		return len(b), nil
	}
	// act
	got, err := pcf.AnalogRead(description)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], ctrlByteOn)
	gobottest.Assert(t, numCallsRead, 2)
	gobottest.Assert(t, got, want)
}

func TestPCF8591DriverAnalogWrite(t *testing.T) {
	// sequence to write the output:
	// * create new value for the control register (ANAON)
	// * write the control register and value
	//
	// arrange
	pcf, adaptor := initTestPCF8591DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	pcf.fromMin = 0
	pcf.fromMax = 255
	pcf.lastCtrlByte = 0x00
	pcf.lastAnaOut = 0x00
	ctrlByteOn := uint8(pcf8591_ANAON)
	want := uint8(0x15)
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// act
	err := pcf.AnalogWrite(int(want))
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, len(adaptor.written), 2)
	gobottest.Assert(t, adaptor.written[0], ctrlByteOn)
	gobottest.Assert(t, adaptor.written[1], want)
}

func TestPCF8591DriverAnalogOutputState(t *testing.T) {
	// sequence to set the state:
	// * create the new value (ctrlByte) for the control register (ANAON)
	// * write the register value
	//
	// arrange
	pcf, adaptor := initTestPCF8591DriverWithStubbedAdaptor()
	for bitState := 0; bitState <= 1; bitState++ {
		adaptor.written = []byte{} // reset writes of Start() and former test
		// arrange some values
		pcf.lastCtrlByte = uint8(0x00)
		wantCtrlByteVal := uint8(pcf8591_ANAON)
		if bitState == 0 {
			pcf.lastCtrlByte = uint8(0xFF)
			wantCtrlByteVal = uint8(0xFF & ^pcf8591_ANAON)
		}
		// act
		err := pcf.AnalogOutputState(bitState == 1)
		// assert
		gobottest.Assert(t, err, nil)
		gobottest.Assert(t, len(adaptor.written), 1)
		gobottest.Assert(t, adaptor.written[0], wantCtrlByteVal)
	}
}

func TestPCF8591Driver_rescaleAI(t *testing.T) {
	// the input scales per default from 0...255
	var tests = map[string]struct {
		desc  string
		toMin int
		toMax int
		input byte
		want  int
	}{
		"single_byte_range_min":    {desc: "s.0", toMin: 0, toMax: 255, input: 0, want: 0},
		"single_byte_range_max":    {desc: "m.1", toMin: 0, toMax: 255, input: 255, want: 255},
		"single_below_min":         {desc: "s.2", toMin: 3, toMax: 121, input: 2, want: 3},
		"single_is_max":            {desc: "s.3", toMin: 5, toMax: 6, input: 255, want: 6},
		"single_upscale":           {desc: "m.0", toMin: 337, toMax: 5337, input: 127, want: 2827},
		"diff_grd_range_min":       {desc: "m.2-3", toMin: -180, toMax: 180, input: 0x80, want: -180},
		"diff_grd_range_minus_one": {desc: "m.2-3", toMin: -180, toMax: 180, input: 0xFF, want: -1},
		"diff_grd_range_max":       {desc: "m.2-3", toMin: -180, toMax: 180, input: 0x7F, want: 180},
		"diff_upscale":             {desc: "t.0-3", toMin: -10, toMax: 1234, input: 0x7F, want: 1234},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange
			mc, _ := PCF8591ParseModeChan(tt.desc)
			pcf := NewPCF8591Driver(newI2cTestAdaptor(), WithPCF8591RescaleInput(mc.channel, tt.toMin, tt.toMax))
			// act
			got := pcf.rescaleAI(tt.input, *mc)
			// assert
			gobottest.Assert(t, got, tt.want)
		})
	}
}

func TestPCF8591Driver_rescaleAO(t *testing.T) {
	// the output scales per default from the given value to 0...255
	var tests = map[string]struct {
		fromMin int
		fromMax int
		input   int
		want    uint8
	}{
		"byte_range_min":           {fromMin: 0, fromMax: 255, input: 0, want: 0},
		"byte_range_max":           {fromMin: 0, fromMax: 255, input: 255, want: 255},
		"signed_percent_range_min": {fromMin: -100, fromMax: 100, input: -100, want: 0},
		"signed_percent_range_mid": {fromMin: -100, fromMax: 100, input: 0, want: 127},
		"signed_percent_range_max": {fromMin: -100, fromMax: 100, input: 100, want: 255},
		"voltage_range_min":        {fromMin: 0, fromMax: 5100, input: 1280, want: 64},
		"upscale":                  {fromMin: 0, fromMax: 24, input: 12, want: 127},
		"below_min":                {fromMin: -10, fromMax: 10, input: -11, want: 0},
		"exceed_max":               {fromMin: 0, fromMax: 20, input: 21, want: 255},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange
			pcf := NewPCF8591Driver(newI2cTestAdaptor(), WithPCF8591RescaleOutput(tt.fromMin, tt.fromMax))
			// act
			got := pcf.rescaleAO(tt.input)
			// assert
			gobottest.Assert(t, got, tt.want)
		})
	}
}
