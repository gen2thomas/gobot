package i2c

import (
	"errors"
	"testing"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/gobottest"
)

var _ gobot.Driver = (*PCA953xDriver)(nil)

var registerTestBytes int

func initPCA953xTestDriver() (*PCA953xDriver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor()
	pca := NewPCA953xDriver(adaptor)
	pca.Start()
	return pca, adaptor
}

func TestPCA953xNewType(t *testing.T) {
	// arrange, act
	var bm interface{} = NewPCA953xDriver(newI2cTestAdaptor())
	// assert
	_, ok := bm.(*PCA953xDriver)
	if !ok {
		t.Errorf("NewPCA953xDriver() should have returned a *PCA953xDriver")
	}
}

func TestPCA953xSetName(t *testing.T) {
	// arrange
	d, _ := initPCA953xTestDriver()
	// act
	d.SetName("NowTestPCA953x")
	// assert
	gobottest.Assert(t, d.Name(), "NowTestPCA953x")
}

func TestPCA953xConnection(t *testing.T) {
	// arrange
	p := NewPCA953xDriver(newI2cTestAdaptor())
	// act, assert
	gobottest.Refute(t, p.Connection(), nil)
}

func TestPCA953xStart(t *testing.T) {
	// arrange
	adaptor := newI2cTestAdaptor()
	pca := NewPCA953xDriver(adaptor)
	// act, assert
	gobottest.Assert(t, pca.Start(), nil)
}

func TestPCA953xStartConnectError(t *testing.T) {
	// arrange
	adaptor := newI2cTestAdaptor()
	adaptor.Testi2cConnectErr(true)
	pca := NewPCA953xDriver(adaptor)
	// act, assert
	gobottest.Assert(t, pca.Start(), errors.New("Invalid i2c connection"))
}

func TestPCA953xHalt(t *testing.T) {
	// arrange
	pca, _ := initPCA953xTestDriver()
	// act, assert
	gobottest.Assert(t, pca.Halt(), nil)
}

func TestPCA953xReadRegister(t *testing.T) {
	// arrange
	const expectedRegAddress = uint8(3)
	const expectedReadByteCount = 1
	var regAddress uint8
	var bytes int

	pca, adaptor := initPCA953xTestDriver()
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		regAddress = b[0]
		return 0, nil
	}
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		bytes = len(b)
		return bytes, nil
	}
	// act
	val, err := pca.readRegister(expectedRegAddress)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, val, uint8(0))
	gobottest.Assert(t, bytes, expectedReadByteCount)
	gobottest.Assert(t, regAddress, expectedRegAddress)
}

func TestPCA953xWriteRegister(t *testing.T) {
	// arrange
	const expectedRegAddress = uint8(0x03)
	const expectedRegVal = uint8(0x97)
	const expectedByteCount = 1
	var regAddress uint8
	var regVal uint8
	var bytesCountAddress int
	var bytesCountVal int

	pca, adaptor := initPCA953xTestDriver()
	numCalls := 0
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		numCalls++
		if numCalls == 1 {
			bytesCountAddress = len(b)
			regAddress = b[0]
			return 0, nil
		}
		if numCalls == 2 {
			bytesCountVal = len(b)
			regVal = b[0]
			return 0, nil
		}		
		return 0, errors.New("to much calls")
	}
	// act
	err := pca.writeRegister(expectedRegAddress, expectedRegVal)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, numCalls, 2)
	gobottest.Assert(t, bytesCountAddress, expectedByteCount)
	gobottest.Assert(t, bytesCountVal, expectedByteCount)
	gobottest.Assert(t, regAddress, expectedRegAddress)
	gobottest.Assert(t, regVal, expectedRegVal)

}

func TestPCA953xWriteClearBit(t *testing.T) {
	// arrange
	pca, adaptor := initPCA953xTestDriver()
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

func TestPCA953xWriteSetBit(t *testing.T) {
	// arrange
	pca, adaptor := initPCA953xTestDriver()
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

func TestPCA953xWriteError(t *testing.T) {
	// arrange
	pca, adaptor := initPCA953xTestDriver()
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

func TestPCA953xRead(t *testing.T) {
	// read
	pca, adaptor := initPCA953xTestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		copy(b, []byte{255})
		return 1, nil
	}
	// act
	val, _ := pca.read()
	// assert
	gobottest.Assert(t, val, uint8(255))
}

func TestPCA953xReadError(t *testing.T) {
	// arrange
	pca, adaptor := initPCA953xTestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), errors.New("read error")
	}
	// act
	val, err := pca.read()
	// assert
	gobottest.Assert(t, val, uint8(0))
	gobottest.Assert(t, err, errors.New("read error"))
}
