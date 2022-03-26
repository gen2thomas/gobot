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

func TestPCF8591DriverAnalogReadSingle(t *testing.T) {
	// sequence to read the input channel:
	// * prepare value (with channel and mode) and write control register
	// * read 3 values to drop (see description in implementation)
	// * read the analog value
	//
	// arrange
	pcf, adaptor := initTestPCF8591DriverWithStubbedAdaptor()
	WithPCF8591RescaleInput(1, 0, 255)(pcf)
	adaptor.written = []byte{} // reset writes of Start() and former test
	description := "s.1"
	pcf.lastCtrlByte = 0x00
	ctrlByteOn := uint8(pcf8591_ALLSINGLE | pcf8591_CHAN1)
	returnRead := []uint8{0x01, 0x02, 0x03, 0xFF}
	want := int32(returnRead[3])
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		if numCallsRead == 1 {
			b = returnRead[0 : len(b)-1]
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
	WithPCF8591RescaleInput(2, -128, 127)(pcf)
	adaptor.written = []byte{} // reset writes of Start() and former test
	description := "m.2-3"
	pcf.lastCtrlByte = 0x00
	ctrlByteOn := uint8(pcf8591_MIXED | pcf8591_CHAN2)
	// some two' complements
	// 0x80 => -128
	// 0xFF => -1
	// 0x00 => 0
	// 0x7F => 127
	returnRead := []uint8{0x01, 0x02, 0x03, 0xFF}
	want := int32(-1)
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		if numCallsRead == 1 {
			b = returnRead[0 : len(b)-1]
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
	WithPCF8591RescaleOutput(0, 255)(pcf)
	adaptor.written = []byte{} // reset writes of Start() and former test
	pcf.lastCtrlByte = 0x00
	pcf.lastAnaOut = 0x00
	ctrlByteOn := uint8(pcf8591_ANAON)
	want := uint8(0x15)
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// act
	err := pcf.AnalogWrite(int32(want))
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
