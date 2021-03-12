package i2c

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
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
)

func initTestPCA9501Driver() (driver *PCA9501Driver) {
	driver, _ = initTestPCA9501DriverWithStubbedAdaptor()
	return
}

func initTestPCA9501DriverWithStubbedAdaptor() (*PCA9501Driver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor()
	return NewPCA9501Driver(adaptor), adaptor
}

func TestNewPCA9501Driver(t *testing.T) {
	var bm interface{} = NewPCA9501Driver(newI2cTestAdaptor())
	_, ok := bm.(*PCA9501Driver)
	if !ok {
		t.Errorf("NewPCA9501Driver() should have returned a *PCA9501Driver")
	}

	p := NewPCA9501Driver(newI2cTestAdaptor())
	gobottest.Refute(t, p.Connection(), nil)
}

func TestPCA9501DriverStart(t *testing.T) {
	pca, _ := initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
}

func TestPCA9501StartConnectError(t *testing.T) {
	d, adaptor := initTestPCA9501DriverWithStubbedAdaptor()
	adaptor.Testi2cConnectErr(true)
	gobottest.Assert(t, d.Start(), errors.New("Invalid i2c connection"))
}

func TestPCA9501DriverHalt(t *testing.T) {
	pca := initTestPCA9501Driver()
	gobottest.Assert(t, pca.Halt(), nil)
}

func TestPCA9501DriverCommandsWriteGPIO(t *testing.T) {
	pca, adaptor := initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	result := pca.Command("WriteGPIO")(pinVal)
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
}

func TestPCA9501DriverCommandsReadGPIO(t *testing.T) {
	pca, adaptor := initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	result := pca.Command("ReadGPIO")(pin)
	gobottest.Assert(t, result.(map[string]interface{})["err"], nil)
}

func TestPCA9501DriverWriteGPIO(t *testing.T) {
	pca, adaptor := initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	err := pca.WriteGPIO(7, 0)
	gobottest.Assert(t, err, nil)
}

func TestPCA9501DriverWriteGPIOErrCTRL(t *testing.T) {
	pca, adaptor := initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}
	err := pca.WriteGPIO(7, 0)
	gobottest.Assert(t, err, errors.New("write error"))
}

func TestPCA9501DriverWriteGPIOErrVAL(t *testing.T) {
	pca, adaptor := initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
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
	err := pca.WriteGPIO(7, 0)
	gobottest.Assert(t, err, errors.New("write error"))
}

func TestPCA9501DriverReadGPIO(t *testing.T) {
	// positive test
	pca, adaptor := initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	val, _ := pca.ReadGPIO(7)
	gobottest.Assert(t, val, uint8(0))
	// error while read
	pca, adaptor = initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), errors.New("error while read")
	}
	_, err := pca.ReadGPIO(7)
	gobottest.Assert(t, err, errors.New("error while read"))
}

func TestPCA9501DriverWrite(t *testing.T) {
	// clear bit
	pca, adaptor := initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	err := pca.write(uint8(7), 0)
	gobottest.Assert(t, err, nil)
	// set bit
	pca, adaptor = initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	err = pca.write(uint8(7), 1)
	gobottest.Assert(t, err, nil)
	// write error
	pca, adaptor = initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)

	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}
	err = pca.write(uint8(7), 0)
	gobottest.Assert(t, err, errors.New("write error"))
	//debug
	debug = true
	log.SetOutput(ioutil.Discard)
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, nil
	}
	err = pca.write(uint8(7), 1)
	gobottest.Assert(t, err, nil)
	debug = false
	log.SetOutput(os.Stdout)
}

func TestPCA9501DriverRead(t *testing.T) {
	// read
	pca, adaptor := initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		copy(b, []byte{255})
		return 1, nil
	}
	val, _ := pca.read()
	gobottest.Assert(t, val, uint8(255))
	// read error
	pca, adaptor = initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)

	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), errors.New("read error")
	}
	val, err := pca.read()
	gobottest.Assert(t, val, uint8(0))
	gobottest.Assert(t, err, errors.New("read error"))
	// debug
	debug = true
	log.SetOutput(ioutil.Discard)
	pca, adaptor = initTestPCA9501DriverWithStubbedAdaptor()
	gobottest.Assert(t, pca.Start(), nil)
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		copy(b, []byte{255})
		return 1, nil
	}
	val, _ = pca.read()
	gobottest.Assert(t, val, uint8(255))
	debug = false
	log.SetOutput(os.Stdout)
}

func TestSetBitAtPos(t *testing.T) {
	var expectedVal uint8 = 129
	actualVal := setBitAtPos(1, 7)
	gobottest.Assert(t, expectedVal, actualVal)
}

func TestClearBitAtPos(t *testing.T) {
	var expectedVal uint8
	actualVal := clearBitAtPos(128, 7)
	gobottest.Assert(t, expectedVal, actualVal)
}

func TestPCA9501DriverSetName(t *testing.T) {
	d := initTestPCA9501Driver()
	d.SetName("TESTME")
	gobottest.Assert(t, d.Name(), "TESTME")
}
