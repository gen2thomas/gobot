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

func TestPCA9501DriverWriteGPIO(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	// act
	err := pca.WriteGPIO(7, 0)
	// assert
	gobottest.Assert(t, err, nil)
}

func TestPCA9501DriverWriteGPIOErrCTRL(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}
	// act
	err := pca.WriteGPIO(7, 0)
	// assert
	gobottest.Assert(t, err, errors.New("write error"))
}

func TestPCA9501DriverWriteGPIOErrVAL(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	numCalls := 1
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		if numCalls == 2 {
			return 0, errors.New("write error")
		}
		numCalls++
		return 0, nil
	}
	// act
	err := pca.WriteGPIO(7, 0)
	// assert
	gobottest.Assert(t, err, errors.New("write error"))
}

func TestPCA9501DriverWriteEEPROM(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	// act
	err := pca.WriteEEPROM(15, 7)
	// assert
	gobottest.Assert(t, err, nil)
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

func TestPCA9501DriverReadGPIOErrorWhileRead(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), errors.New("error while read")
	}
	// act
	_, err := pca.ReadGPIO(7)
	// assert
	gobottest.Assert(t, err, errors.New("error while read"))
}

func TestPCA9501DriverReadEEPROM(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// act
	val, _ := pca.ReadEEPROM(15)
	// assert
	gobottest.Assert(t, val, uint8(0))
}

func TestPCA9501DriverReadEEPROMErrorWhileRead(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), errors.New("error while read")
	}
	// act
	_, err := pca.ReadEEPROM(15)
	// assert
	gobottest.Assert(t, err, errors.New("error while read"))
}

func TestPCA9501DriverWriteClearBit(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	// act
	err := pca.write(uint8(0))
	// assert
	gobottest.Assert(t, err, nil)
}

func TestPCA9501DriverWriteSetBit(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	// act
	err := pca.write(uint8(7))
	// assert
	gobottest.Assert(t, err, nil)
}

func TestPCA9501DriverWriteError(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}
	// act
	err := pca.write(uint8(7))
	// assert
	gobottest.Assert(t, err, errors.New("write error"))
}

func TestPCA9501DriverRead(t *testing.T) {
	// read
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		copy(b, []byte{255})
		return 1, nil
	}
	// act
	val, _ := pca.read()
	// assert
	gobottest.Assert(t, val, uint8(255))
}

func TestPCA9501DriverReadError(t *testing.T) {
	// arrange
	pca, adaptor := initPCA9501TestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), errors.New("read error")
	}
	// act
	val, err := pca.read()
	// assert
	gobottest.Assert(t, val, uint8(0))
	gobottest.Assert(t, err, errors.New("read error"))
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
