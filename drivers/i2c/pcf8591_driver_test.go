package i2c

import (
	"testing"

	"gobot.io/x/gobot/gobottest"
)

func initTestPCF8591DriverWithStubbedAdaptor() (*PCF8591Driver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor()
	pcf := NewPCF8591Driver(adaptor)
	pcf.Start()
	return pcf, adaptor
}

func TestPCF8591DriverAnalogReadSingle(t *testing.T) {
	// sequence to read the input channel:
	// * prepare value (with channel and config) and write control register
	// * read the analog value
	//
	// TODO: there seems to be no possibility to read the control register
	// so possibly we have to store it by our own
	// TODO: it is possible to read more than one value without repeat the
	// start and control byte process until next write or read from another channel.
	// So, possibly we can speedup reading by compare the last written controlbyte?
	//
	// arrange
	pcf, adaptor := initTestPCF8591DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	description := "s.1"
	wantCtrlByteVal := uint8(pcf8591_ANAON | pcf8591_ALLSINGLE | pcf8591_CHAN1)
	returnRead := uint8(0xFF)
	want := int(returnRead)
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[0] = returnRead
		return len(b), nil
	}
	// act
	got, err := pcf.AnalogRead(description)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], wantCtrlByteVal)
	gobottest.Assert(t, numCallsRead, 1)
	gobottest.Assert(t, got, want)
}

func TestPCF8591DriverAnalogReadDiff(t *testing.T) {
	// sequence to read the input channel:
	// * prepare value (with channel and config) and write control register
	// * read the analog value
	// * convert to 8-bit two's complement (-127...128)
	//
	// TODO: there seems to be no possibility to read the control register
	// so possibly we have to store it by our own
	// TODO: it is possible to read more than one value without repeat the
	// start and control byte process until next write or read from another channel.
	// So, possibly we can speedup reading by compare the last written controlbyte?
	//
	// arrange
	pcf, adaptor := initTestPCF8591DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	description := "m.2-3"
	wantCtrlByteVal := uint8(pcf8591_ANAON | pcf8591_MIXED | pcf8591_CHAN2)
	returnRead := uint8(0xFF)
	want := int(returnRead) - 128
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[0] = uint8(returnRead)
		return len(b), nil
	}
	// act
	got, err := pcf.AnalogRead(description)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], wantCtrlByteVal)
	gobottest.Assert(t, numCallsRead, 1)
	gobottest.Assert(t, got, want)
}

func TestPCF8591DriverAnalogWrite(t *testing.T) {
	// sequence to write the output:
	// * write ANAON to the control register => done by AnalogOutputState(true)
	// * write the analog value
	//
	// TODO: there seems to be no possibility to read the control register
	// so possibly we have to store it by our own
	// TODO: it is possible to write more than one analog value without repeat the
	// start and control byte process until next read. So, possibly we can speedup
	// writing with an boolean marker (e.g. writePrepared) or simply compare the controlbyte?
	//
	// arrange
	pcf, adaptor := initTestPCF8591DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	wantCtrlByteVal := uint8(pcf8591_DAMASK | pcf8591_ANAON)
	want := uint8(0x15)
	// act
	err := pcf.AnalogWrite(want)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, len(adaptor.written), 2)
	gobottest.Assert(t, adaptor.written[0], wantCtrlByteVal)
	gobottest.Assert(t, adaptor.written[1], want)
}

func TestPCF8591DriverAnalogOutputState(t *testing.T) {
	// sequence to set the state:
	// * create the new value (ctrlByte) of the control register (ANAOFF, ANAON)
	// * write the register value
	//
	// TODO: there seems to be no possibility to read the control register
	// so possibly we have to store it by our own
	//
	// arrange
	pcf, adaptor := initTestPCF8591DriverWithStubbedAdaptor()
	ctrlByteOn := uint8(pcf8591_DAMASK | pcf8591_ANAON)
	ctrlByteOff := uint8(pcf8591_DAMASK & ^pcf8591_ANAON)
	for bitState := 0; bitState <= 1; bitState++ {
		adaptor.written = []byte{} // reset writes of Start() and former test
		// arrange some values
		wantCtrlByteVal := ctrlByteOn
		if bitState == 0 {
			wantCtrlByteVal = ctrlByteOff
		}
		// act
		err := pcf.AnalogOutputState(bitState == 1)
		// assert
		gobottest.Assert(t, err, nil)
		gobottest.Assert(t, len(adaptor.written), 1)
		gobottest.Assert(t, adaptor.written[0], wantCtrlByteVal)
	}
}
