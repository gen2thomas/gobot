package i2c

import (
	"strings"
	"testing"
	"time"

	"gobot.io/x/gobot/gobottest"
)

func initTestPCF8583DriverWithStubbedAdaptor() (*PCF8583Driver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor()
	pcf := NewPCF8583Driver(adaptor)
	pcf.Start()
	return pcf, adaptor
}

func TestNewPCF8583Driver(t *testing.T) {
	// arrange
	adaptor := newI2cTestAdaptor()
	//act
	pcf := NewPCF8583Driver(adaptor)
	//assert
	gobottest.Assert(t, pcf.connector, adaptor)
	gobottest.Assert(t, pcf.mode, PCF8583Control(0x00))
	gobottest.Assert(t, pcf.ramOffset, uint8(0x10))
	gobottest.Refute(t, pcf.mutex, nil)
}

func TestNewPCF8583DriverWithPCF8583Mode(t *testing.T) {
	// arrange
	adaptor := newI2cTestAdaptor()
	pcf := NewPCF8583Driver(adaptor)
	pcf.mode = PCF8583ModeTest
	want := PCF8583ModeClock50
	//act
	WithPCF8583Mode(want)(pcf)
	//assert
	gobottest.Assert(t, pcf.mode, want)
}

func TestPCF8583DriverStartNoModeSwitch(t *testing.T) {
	// arrange
	adaptor := newI2cTestAdaptor()
	pcf := NewPCF8583Driver(adaptor)
	adaptor.written = []byte{}   // reset writes of former tests
	readCtrlState := uint8(0x01) // 32.768kHz clock mode
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[len(b)-1] = readCtrlState
		return len(b), nil
	}
	// act
	err := pcf.Start()
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, numCallsRead, 1)
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], uint8(pcf8583_CTRL))
	gobottest.Refute(t, pcf.connection, nil)
}

func TestPCF8583DriverStartWithModeSwitch(t *testing.T) {
	// sequence to change mode:
	// * read control register for get current state
	// * reset old mode bits and set new mode bit
	// * write the control register
	// arrange
	adaptor := newI2cTestAdaptor()
	pcf := NewPCF8583Driver(adaptor)
	pcf.mode = PCF8583ModeCounter
	adaptor.written = []byte{}   // reset writes of former tests
	readCtrlState := uint8(0x02) // 32.768kHz clock mode
	wantReg0Val := uint8(0x22)   // event counter mode
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[len(b)-1] = readCtrlState
		return len(b), nil
	}
	// act
	err := pcf.Start()
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, numCallsRead, 1)
	gobottest.Assert(t, len(adaptor.written), 3)
	gobottest.Assert(t, adaptor.written[0], uint8(pcf8583_CTRL))
	gobottest.Assert(t, adaptor.written[1], uint8(pcf8583_CTRL))
	gobottest.Assert(t, adaptor.written[2], uint8(wantReg0Val))
}

func TestPCF8583DriverWriteTime(t *testing.T) {
	// sequence to write the time:
	// * read control register for get current state and ensure an clock mode is set
	// * write the control register (stop counting)
	// * create the values for date registers (default is 24h mode)
	// * write the clock and calendar registers with auto increment
	// * write the control register (start counting)
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{}         // reset writes of Start() and former test
	readCtrlState := uint8(0x07)       // 32.768kHz clock mode
	milliSec := 210 * time.Millisecond // 0.21 sec = 210 ms
	initDate := time.Date(2022, time.December, 16, 15, 14, 13, int(milliSec), time.UTC)
	wantCtrlStop := uint8(0x87)  // stop counting bit is set
	wantReg1Val := uint8(0x21)   // BCD: 1/10 and 1/100 sec (21)
	wantReg2Val := uint8(0x13)   // BCD: 10 and 1 sec (13)
	wantReg3Val := uint8(0x14)   // BCD: 10 and 1 min (14)
	wantReg4Val := uint8(0x15)   // BCD: 10 and 1 hour (15)
	wantReg5Val := uint8(0x16)   // year (0) and BCD: date (16)
	wantReg6Val := uint8(0xB2)   // weekday 5, bit 5 and bit 7 (0xA0) and BCD: month (0x12)
	wantCrtlStart := uint8(0x07) // stop counting bit is reset
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[len(b)-1] = readCtrlState
		return len(b), nil
	}
	// act
	err := pcf.WriteTime(initDate)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, pcf.yearOffset, initDate.Year())
	gobottest.Assert(t, numCallsRead, 1)
	gobottest.Assert(t, len(adaptor.written), 11)
	gobottest.Assert(t, adaptor.written[0], uint8(pcf8583_CTRL))
	gobottest.Assert(t, adaptor.written[1], uint8(pcf8583_CTRL))
	gobottest.Assert(t, adaptor.written[2], wantCtrlStop)
	gobottest.Assert(t, adaptor.written[3], wantReg1Val)
	gobottest.Assert(t, adaptor.written[4], wantReg2Val)
	gobottest.Assert(t, adaptor.written[5], wantReg3Val)
	gobottest.Assert(t, adaptor.written[6], wantReg4Val)
	gobottest.Assert(t, adaptor.written[7], wantReg5Val)
	gobottest.Assert(t, adaptor.written[8], wantReg6Val)
	gobottest.Assert(t, adaptor.written[9], uint8(pcf8583_CTRL))
	gobottest.Assert(t, adaptor.written[10], wantCrtlStart)
}

func TestPCF8583DriverWriteTimeNoTimeModeFails(t *testing.T) {
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{}   // reset writes of Start() and former test
	readCtrlState := uint8(0x30) // test mode
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[len(b)-1] = readCtrlState
		return len(b), nil
	}
	// act
	err := pcf.WriteTime(time.Now())
	// assert
	gobottest.Refute(t, err, nil)
	gobottest.Assert(t, strings.Contains(err.Error(), "wrong mode 0x30"), true)
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], uint8(pcf8583_CTRL))
	gobottest.Assert(t, numCallsRead, 1)
}

func TestPCF8583DriverReadTime(t *testing.T) {
	// sequence to read the time:
	// * read the control register to determine mask flag and ensure an clock mode is set
	// * read the clock and calendar registers with auto increment
	// * create the value out of registers content
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	pcf.yearOffset = 2020
	milliSec := 210 * time.Millisecond // 0.21 sec = 210 ms
	want := time.Date(2022, time.December, 16, 15, 14, 13, int(milliSec), time.UTC)
	reg0Val := uint8(0x10) // clock mode 50Hz
	reg1Val := uint8(0x21) // BCD: 1/10 and 1/100 sec (21)
	reg2Val := uint8(0x13) // BCD: 10 and 1 sec (13)
	reg3Val := uint8(0x14) // BCD: 10 and 1 min (14)
	reg4Val := uint8(0x15) // BCD: 10 and 1 hour (15)
	reg5Val := uint8(0x96) // year (2) and BCD: date (16)
	reg6Val := uint8(0xB2) // weekday 5, bit 5 and bit 7 (0xA0) and BCD: month (0x12)
	returnRead := [2][]uint8{
		{reg0Val},
		{reg1Val, reg2Val, reg3Val, reg4Val, reg5Val, reg6Val},
	}
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		rr := returnRead[numCallsRead-1]
		for i := 0; i < len(b); i++ {
			b[i] = rr[i]
		}
		return len(b), nil
	}
	// act
	got, err := pcf.ReadTime()
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], uint8(pcf8583_CTRL))
	gobottest.Assert(t, numCallsRead, 2)
	gobottest.Assert(t, got, want)
}

func TestPCF8583DriverReadTimeNoTimeModeFails(t *testing.T) {
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{}   // reset writes of Start() and former test
	readCtrlState := uint8(0x20) // counter mode
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[len(b)-1] = readCtrlState
		return len(b), nil
	}
	// act
	got, err := pcf.ReadTime()
	// assert
	gobottest.Refute(t, err, nil)
	gobottest.Assert(t, strings.Contains(err.Error(), "wrong mode 0x20"), true)
	gobottest.Assert(t, got, time.Time{})
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], uint8(pcf8583_CTRL))
	gobottest.Assert(t, numCallsRead, 1)
}

func TestPCF8583DriverWriteCounter(t *testing.T) {
	// sequence to write the counter:
	// * read control register for get current state and ensure the event counter mode is set
	// * write the control register (stop counting)
	// * create the values for counter registers
	// * write the counter registers
	// * write the control register (start counting)
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{}   // reset writes of Start() and former test
	readCtrlState := uint8(0x27) // counter mode
	initCount := int32(654321)   // 6 digits used of 10 possible with int32
	wantCtrlStop := uint8(0xA7)  // stop counting bit is set
	wantReg1Val := uint8(0x21)   // BCD: 21
	wantReg2Val := uint8(0x43)   // BCD: 43
	wantReg3Val := uint8(0x65)   // BCD: 65
	wantCtrlStart := uint8(0x27) // counter mode
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[len(b)-1] = readCtrlState
		return len(b), nil
	}
	// act
	err := pcf.WriteCounter(initCount)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, numCallsRead, 1)
	gobottest.Assert(t, len(adaptor.written), 8)
	gobottest.Assert(t, adaptor.written[0], uint8(pcf8583_CTRL))
	gobottest.Assert(t, adaptor.written[1], uint8(pcf8583_CTRL))
	gobottest.Assert(t, adaptor.written[2], wantCtrlStop)
	gobottest.Assert(t, adaptor.written[3], wantReg1Val)
	gobottest.Assert(t, adaptor.written[4], wantReg2Val)
	gobottest.Assert(t, adaptor.written[5], wantReg3Val)
	gobottest.Assert(t, adaptor.written[6], uint8(pcf8583_CTRL))
	gobottest.Assert(t, adaptor.written[7], wantCtrlStart)
}

func TestPCF8583DriverWriteCounterNoCounterModeFails(t *testing.T) {
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{}   // reset writes of Start() and former test
	readCtrlState := uint8(0x10) // 50Hz mode
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[len(b)-1] = readCtrlState
		return len(b), nil
	}
	// act
	err := pcf.WriteCounter(123)
	// assert
	gobottest.Refute(t, err, nil)
	gobottest.Assert(t, strings.Contains(err.Error(), "wrong mode 0x10"), true)
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], uint8(pcf8583_CTRL))
	gobottest.Assert(t, numCallsRead, 1)
}

func TestPCF8583DriverReadCounter(t *testing.T) {
	// sequence to read the counter:
	// * read the control register to ensure the event counter mode is set
	// * read the counter registers
	// * create the value out of registers content
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	want := int32(654321)
	reg0Val := uint8(0x20) // counter mode
	reg1Val := uint8(0x21) // BCD: 21
	reg2Val := uint8(0x43) // BCD: 43
	reg3Val := uint8(0x65) // BCD: 65
	returnRead := [2][]uint8{
		{reg0Val},
		{reg1Val, reg2Val, reg3Val},
	}
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		rr := returnRead[numCallsRead-1]
		for i := 0; i < len(b); i++ {
			b[i] = rr[i]
		}
		return len(b), nil
	}
	// act
	got, err := pcf.ReadCounter()
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], uint8(pcf8583_CTRL))
	gobottest.Assert(t, numCallsRead, 2)
	gobottest.Assert(t, got, want)
}

func TestPCF8583DriverReadCounterNoCounterModeFails(t *testing.T) {
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{}   // reset writes of Start() and former test
	readCtrlState := uint8(0x30) // test mode
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[len(b)-1] = readCtrlState
		return len(b), nil
	}
	// act
	got, err := pcf.ReadCounter()
	// assert
	gobottest.Refute(t, err, nil)
	gobottest.Assert(t, strings.Contains(err.Error(), "wrong mode 0x30"), true)
	gobottest.Assert(t, got, int32(0))
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], uint8(pcf8583_CTRL))
	gobottest.Assert(t, numCallsRead, 1)
}

func TestPCF8583DriverWriteRam(t *testing.T) {
	// sequence to write the RAM:
	// * calculate the RAM address and check for valid range
	// * write the given value to the given RAM address
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	wantRamAddress := uint8(0xFF)
	wantRamValue := uint8(0xEF)
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// act
	err := pcf.WriteRAM(wantRamAddress-pcf8583RamOffset, wantRamValue)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, len(adaptor.written), 2)
	gobottest.Assert(t, adaptor.written[0], wantRamAddress)
	gobottest.Assert(t, adaptor.written[1], wantRamValue)
}

func TestPCF8583DriverWriteRamAddressOverflowFails(t *testing.T) {
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	// act
	err := pcf.WriteRAM(uint8(0xF0), 15)
	// assert
	gobottest.Refute(t, err, nil)
	gobottest.Assert(t, strings.Contains(err.Error(), "overflow 256"), true)
	gobottest.Assert(t, len(adaptor.written), 0)
}

func TestPCF8583DriverReadRam(t *testing.T) {
	// sequence to read the RAM:
	// * calculate the RAM address and check for valid range
	// * read the value from the given RAM address
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	wantRamAddress := uint8(pcf8583RamOffset)
	want := uint8(0xAB)
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[len(b)-1] = want
		return len(b), nil
	}
	// act
	got, err := pcf.ReadRAM(wantRamAddress - pcf8583RamOffset)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, got, want)
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], wantRamAddress)
	gobottest.Assert(t, numCallsRead, 1)
}

func TestPCF8583DriverReadRamAddressOverflowFails(t *testing.T) {
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	adaptor.written = []byte{} // reset writes of Start() and former test
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		return len(b), nil
	}
	// act
	got, err := pcf.ReadRAM(uint8(0xF0))
	// assert
	gobottest.Refute(t, err, nil)
	gobottest.Assert(t, strings.Contains(err.Error(), "overflow 256"), true)
	gobottest.Assert(t, got, uint8(0))
	gobottest.Assert(t, len(adaptor.written), 0)
	gobottest.Assert(t, numCallsRead, 0)
}

func TestPCF8583DriverCommandsWriteTime(t *testing.T) {
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	readCtrlState := uint8(0x10) // clock 50Hz
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[len(b)-1] = readCtrlState
		return len(b), nil
	}
	// act
	result := pcf.Command("WriteTime")(map[string]interface{}{"val": time.Now()})
	// assert
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
}

func TestPCF8583DriverCommandsReadTime(t *testing.T) {
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	pcf.yearOffset = 2019
	milliSec := 550 * time.Millisecond // 0.55 sec = 550 ms
	want := time.Date(2021, time.December, 24, 18, 00, 00, int(milliSec), time.UTC)
	reg0Val := uint8(0x00) // clock mode 32.768 kHz
	reg1Val := uint8(0x55) // BCD: 1/10 and 1/100 sec (55)
	reg2Val := uint8(0x00) // BCD: 10 and 1 sec (00)
	reg3Val := uint8(0x00) // BCD: 10 and 1 min (00)
	reg4Val := uint8(0x18) // BCD: 10 and 1 hour (18)
	reg5Val := uint8(0xA4) // year (2) and BCD: date (24)
	reg6Val := uint8(0xB2) // weekday 5, bit 5 and bit 7 (0xA0) and BCD: month (0x12)
	returnRead := [2][]uint8{
		{reg0Val},
		{reg1Val, reg2Val, reg3Val, reg4Val, reg5Val, reg6Val},
	}
	// arrange reads
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		rr := returnRead[numCallsRead-1]
		for i := 0; i < len(b); i++ {
			b[i] = rr[i]
		}
		return len(b), nil
	}
	// act
	result := pcf.Command("ReadTime")(map[string]interface{}{})
	// assert
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
	gobottest.Assert(t, result.(map[string]interface{})["val"], want)
}

func TestPCF8583DriverCommandsWriteCounter(t *testing.T) {
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	readCtrlState := uint8(0x20) // counter
	// arrange writes
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[len(b)-1] = readCtrlState
		return len(b), nil
	}
	// act
	result := pcf.Command("WriteCounter")(map[string]interface{}{"val": int32(123456)})
	// assert
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
}

func TestPCF8583DriverCommandsReadCounter(t *testing.T) {
	// arrange
	pcf, adaptor := initTestPCF8583DriverWithStubbedAdaptor()
	want := int32(123456)
	reg0Val := uint8(0x20) // counter mode
	reg1Val := uint8(0x56) // BCD: 56
	reg2Val := uint8(0x34) // BCD: 34
	reg3Val := uint8(0x12) // BCD: 12
	returnRead := [2][]uint8{
		{reg0Val},
		{reg1Val, reg2Val, reg3Val},
	}
	// arrange reads
	// arrange reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		rr := returnRead[numCallsRead-1]
		for i := 0; i < len(b); i++ {
			b[i] = rr[i]
		}
		return len(b), nil
	}
	// act
	result := pcf.Command("ReadCounter")(map[string]interface{}{})
	// assert
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
	gobottest.Assert(t, result.(map[string]interface{})["val"], want)
}

func TestPCF8583DriverCommandsWriteRAM(t *testing.T) {
	// arrange
	pcf, _ := initTestPCF8583DriverWithStubbedAdaptor()
	var addressValue = map[string]interface{}{
		"address": uint8(0x12),
		"val":     uint8(0x45),
	}
	// act
	result := pcf.Command("WriteRAM")(addressValue)
	// assert
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
}

func TestPCF8583DriverCommandsReadRAM(t *testing.T) {
	// arrange
	pcf, _ := initTestPCF8583DriverWithStubbedAdaptor()
	var address = map[string]interface{}{
		"address": uint8(0x34),
	}
	// act
	result := pcf.Command("ReadRAM")(address)
	// assert
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
	gobottest.Assert(t, result.(map[string]interface{})["val"], uint8(0))
}
