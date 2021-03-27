package i2c

import (
	"errors"
	"testing"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/gobottest"
)

var _ gobot.Driver = (*PCA9501Driver)(nil)
var (
	pinVal = map[string]interface{}{
		"pin": uint8(7),
		"val": uint8(0),
	}
	pin = map[string]interface{}{
		"pin": uint8(7),
	}
	addressVal = map[string]interface{}{
		"address": uint8(15),
		"val":     uint8(7),
	}
	address = map[string]interface{}{
		"address": uint8(15),
	}
)

func initPCA9501TestDriver() (*PCA9501Driver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor()
	pca := NewPCA9501Driver(adaptor)
	pca.Start()
	return pca, adaptor
}

func TestPCA9501DriverNewType(t *testing.T) {
	// arrange, act
	var bm interface{} = NewPCA9501Driver(newI2cTestAdaptor())
	// assert
	_, ok := bm.(*PCA9501Driver)
	if !ok {
		t.Errorf("NewPCA9501Driver() should have returned a *PCA9501Driver")
	}
}

func TestPCA9501DriverConnection(t *testing.T) {
	// arrange
	p := NewPCA9501Driver(newI2cTestAdaptor())
	// act, assert
	gobottest.Refute(t, p.Connection(), nil)
}

func TestPCA9501DriverStart(t *testing.T) {
	// arrange
	adaptor := newI2cTestAdaptor()
	pca := NewPCA9501Driver(adaptor)
	// act, assert
	gobottest.Assert(t, pca.Start(), nil)
}

func TestPCA9501DriverStartConnectError(t *testing.T) {
	// arrange
	adaptor := newI2cTestAdaptor()
	adaptor.Testi2cConnectErr(true)
	pca := NewPCA9501Driver(adaptor)
	// act, assert
	gobottest.Assert(t, pca.Start(), errors.New("Invalid i2c connection"))
}

func TestPCA9501DriverHalt(t *testing.T) {
	// arrange
	pca, _ := initPCA9501TestDriver()
	// act, assert
	gobottest.Assert(t, pca.Halt(), nil)
}

func TestPCA9501DriverCommandsWriteGPIO(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	// act
	result := pca.Command("WriteGPIO")(pinVal)
	// assert
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
}

func TestPCA9501DriverCommandsReadGPIO(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// act
	result := pca.Command("ReadGPIO")(pin)
	// assert
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
}

func TestPCA9501DriverCommandsWriteEEPROM(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	// act
	result := pca.Command("WriteEEPROM")(addressVal)
	// assert
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
}

func TestPCA9501DriverCommandsReadEEPROM(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// act
	result := pca.Command("ReadEEPROM")(address)
	// assert
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
}

func TestPCA9501DriverWriteGPIOClearBit(t *testing.T) {
	// arrange
	pinUnderTest := uint8(6)
	pca, adaptor := initPCA9501TestDriver()
	// prepare all reads
	const ioDirAllInput = 0xF1
	const ioStateAllInput = 0xF2
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		if numCallsRead == 1 {
			// first call read current io direction of all pins
			b[0] = ioDirAllInput
		}
		if numCallsRead == 2 {
			// second call read current state of all pins
			b[0] = ioStateAllInput
		}
		return len(b), nil
	}
	// prepare all writes
	const ioDirPinUnderTestExpected = uint8(0xB1)
	const ioStatePinUnderTestExpected = uint8(0xB2)
	numCallsWrite := 0
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		numCallsWrite++
		// first call write io direction with pin under test changed to output (adaptor.written[0])
		// second call write io state with pin under test reset (adaptor.written[1])
		return 0, nil
	}
	// act
	err := pca.WriteGPIO(pinUnderTest, 0)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, numCallsRead, 2)
	gobottest.Assert(t, numCallsWrite, 2)
	gobottest.Assert(t, adaptor.written[0], ioDirPinUnderTestExpected)
	gobottest.Assert(t, adaptor.written[1], ioStatePinUnderTestExpected)
}

func TestPCA9501DriverWriteGPIOSetBit(t *testing.T) {
	// arrange
	pinUnderTest := uint8(3)
	pca, adaptor := initPCA9501TestDriver()
	// prepare all reads
	const ioDirAllInput = 0x1F
	const ioStateAllInput = 0x20
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		if numCallsRead == 1 {
			// first call read current io direction of all pins
			b[0] = ioDirAllInput
		}
		if numCallsRead == 2 {
			// second call read current state of all pins
			b[0] = ioStateAllInput
		}
		return len(b), nil
	}
	// prepare all writes
	const ioDirPinUnderTestExpected = uint8(0x17)
	const ioStatePinUnderTestExpected = uint8(0x28)
	numCallsWrite := 0
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		numCallsWrite++
		// first call write io direction with pin under test changed to output (adaptor.written[0])
		// second call write io state with pin under test reset (adaptor.written[1])
		return 0, nil
	}
	// act
	err := pca.WriteGPIO(pinUnderTest, 2)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, numCallsRead, 2)
	gobottest.Assert(t, numCallsWrite, 2)
	gobottest.Assert(t, adaptor.written[0], ioDirPinUnderTestExpected)
	gobottest.Assert(t, adaptor.written[1], ioStatePinUnderTestExpected)
}

func TestPCA9501DriverWriteGPIOErrorAtWriteDirection(t *testing.T) {
	// arrange
	expectedWriteError := errors.New("write error")
	pca, adaptor := initPCA9501TestDriver()
	// prepare all reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		return len(b), nil
	}
	// prepare all writes
	numCallsWrite := 0
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		numCallsWrite++
		if numCallsWrite == 1 {
			// first call writes the CTRL register for port direction
			return 0, expectedWriteError
		}
		return 0, nil
	}
	// act
	err := pca.WriteGPIO(7, 0)
	// assert
	gobottest.Assert(t, err, expectedWriteError)
	gobottest.Assert(t, numCallsRead < 2, true)
	gobottest.Assert(t, numCallsWrite, 1)
}

func TestPCA9501DriverWriteGPIOErrorAtWriteValue(t *testing.T) {
	// arrange
	expectedWriteError := errors.New("write error")
	pca, adaptor := initPCA9501TestDriver()
	// prepare all reads
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// prepare all writes
	numCallsWrite := 0
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		numCallsWrite++
		if numCallsWrite == 2 {
			// second call writes the value to IO port
			return 0, expectedWriteError
		}
		return 0, nil
	}
	// act
	err := pca.WriteGPIO(7, 0)
	// assert
	gobottest.Assert(t, err, expectedWriteError)
	gobottest.Assert(t, numCallsWrite, 2)
}

func TestPCA9501DriverReadGPIO(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// act
	val, _ := pca.ReadGPIO(7)
	// assert
	gobottest.Assert(t, val, uint8(0))
}

func TestPCA9501DriverReadGPIOErrorAtReadDirection(t *testing.T) {
	// arrange
	expectedReadError := errors.New("read error")
	pca, adaptor := initPCA9501TestDriver()
	// prepare all reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		if numCallsRead == 1 {
			// first read gets the CTRL register for pin direction
			return 0, expectedReadError
		}
		return len(b), nil
	}
	// prepare all writes
	numCallsWrite := 0
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		numCallsWrite++
		return 0, nil
	}
	// act
	_, err := pca.ReadGPIO(1)
	// assert
	gobottest.Assert(t, err, expectedReadError)
	gobottest.Assert(t, numCallsRead, 1)
	gobottest.Assert(t, numCallsWrite, 0)
}

func TestPCA9501DriverReadGPIOErrorAtReadValue(t *testing.T) {
	// arrange
	expectedReadError := errors.New("read error")
	pca, adaptor := initPCA9501TestDriver()
	// prepare all reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		if numCallsRead == 2 {
			// second read gets the value from IO port
			return 0, expectedReadError
		}
		return len(b), nil
	}
	// prepare all writes
	numCallsWrite := 0
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		numCallsWrite++
		return 0, nil
	}
	// act
	_, err := pca.ReadGPIO(2)
	// assert
	gobottest.Assert(t, err, expectedReadError)
	gobottest.Assert(t, numCallsWrite, 1)
}

func TestPCA9501DriverWriteEEPROM(t *testing.T) {
	// arrange
	const addressUnderTest = uint8(0x52)
	const valueUnderTest = uint8(0x25)
	pca, adaptor := initPCA9501TestDriver()
	// prepare all writes
	numCallsWrite := 0
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		numCallsWrite++
		return 0, nil
	}
	// act
	err := pca.WriteEEPROM(addressUnderTest, valueUnderTest)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, numCallsWrite, 1)
	gobottest.Assert(t, adaptor.written[0], addressUnderTest)
	gobottest.Assert(t, adaptor.written[1], valueUnderTest)
}

func TestPCA9501DriverWriteEEPROMWithDummyAddressReturnsError(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	dummy := pca.MemReadDummy()
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	// act
	err := pca.WriteEEPROM(dummy.Address, 7)
	// assert
	gobottest.Assert(t, err != nil, true)
}

func TestPCA9501DriverReadEEPROM(t *testing.T) {
	// arrange
	addressUnderTest := uint8(51)
	pca, adaptor := initPCA9501TestDriver()
	// prepare all writes
	dummy := pca.MemReadDummy()
	numCallsWrite := 0
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		numCallsWrite++
		return 0, nil
	}
	// prepare all reads
	expectedVal := uint8(0x44)
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		b[addressUnderTest-1-dummy.Address] = expectedVal
		return len(b), nil
	}
	// act
	val, _ := pca.ReadEEPROM(addressUnderTest)
	// assert
	gobottest.Assert(t, val, expectedVal)
	gobottest.Assert(t, numCallsWrite, 1)
	gobottest.Assert(t, adaptor.written[0], dummy.Address)
	gobottest.Assert(t, adaptor.written[1], dummy.Value)
	gobottest.Assert(t, numCallsRead, 1)
}

func TestPCA9501DriverReadEEPROMWithDummyAddressReturnsErrorAndDummyValue(t *testing.T) {
	// arrange
	pca, _ := initPCA9501TestDriver()
	dummy := pca.MemReadDummy()
	// act
	val, err := pca.ReadEEPROM(dummy.Address)
	// assert
	gobottest.Assert(t, err != nil, true)
	gobottest.Assert(t, val, dummy.Value)
}

func TestPCA9501DriverReadEEPROMErrorWhileWriteAddress(t *testing.T) {
	// arrange
	expectedWriteError := errors.New("error while write")
	pca, adaptor := initPCA9501TestDriver()
	// prepare all writes
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, expectedWriteError
	}
	// prepare all reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		return len(b), nil
	}
	// act
	_, err := pca.ReadEEPROM(15)
	// assert
	gobottest.Assert(t, err, expectedWriteError)
	gobottest.Assert(t, numCallsRead, 0)
}

func TestPCA9501DriverReadEEPROMErrorWhileReadValue(t *testing.T) {
	// arrange
	expectedReadError := errors.New("error while read")
	pca, adaptor := initPCA9501TestDriver()
	// prepare all writes
	numCallsWrite := 0
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		numCallsWrite++
		return 0, nil
	}
	// prepare all reads
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), expectedReadError
	}
	// act
	_, err := pca.ReadEEPROM(15)
	// assert
	gobottest.Assert(t, numCallsWrite, 1)
	gobottest.Assert(t, err, expectedReadError)
}

func TestPCA9501DriverSetBitAtPos(t *testing.T) {
	// arrange
	var expectedVal uint8 = 129
	// act
	actualVal := setBitAtPos(1, 7)
	// assert
	gobottest.Assert(t, expectedVal, actualVal)
}

func TestPCA9501DriverClearBitAtPos(t *testing.T) {
	// arrange
	var expectedVal uint8
	// act
	actualVal := clearBitAtPos(128, 7)
	// assert
	gobottest.Assert(t, expectedVal, actualVal)
}

func TestPCA9501DriverSetName(t *testing.T) {
	// arrange
	d, _ := initPCA9501TestDriver()
	// act
	d.SetName("TESTME")
	// assert
	gobottest.Assert(t, d.Name(), "TESTME")
}

func TestPCA9501DriverGetAddressMem(t *testing.T) {
	// arrange
	var expectedVal int = 0x44
	d, _ := initPCA9501TestDriver()
	// act
	actualVal := d.getAddressMem(0x04)
	// assert
	gobottest.Assert(t, expectedVal, actualVal)
}
