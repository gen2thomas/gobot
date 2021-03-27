package i2c

import (
	"errors"
	"math"
	"testing"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/gobottest"
)

var _ gobot.Driver = (*PCA953xDriver)(nil)

type pca953xCalcPscTest struct {
	period      float32
	expectedVal uint8
	expectedErr error
}

type pca953xCalcPeriodTest struct {
	psc         uint8
	expectedVal float32
}

type pca953xCalcPwmTest struct {
	percent     float32
	expectedVal uint8
	expectedErr error
}

type pca953xCalcDutyCycleTest struct {
	pwm         uint8
	expectedVal float32
}

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

func TestPCA953xReadGPIO(t *testing.T) {
	// arrange
	const expectedRegAddress = uint8(0x00) // input register
	const expectedReadByteCount = 1
	var regAddress uint8
	var bytes int

	pca, adaptor := initPCA953xTestDriver()
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		regAddress = b[0]
		bytes = len(b)
		return bytes, nil
	}
	// act
	val, err := pca.ReadGPIO(2) // index doesn't matter
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, val, uint8(0))
	gobottest.Assert(t, bytes, expectedReadByteCount)
	gobottest.Assert(t, regAddress, expectedRegAddress)
}

func TestPCA953xReadGPIOErrorWhileRead(t *testing.T) {
	// arrange
	expectedReadError := errors.New("read error")
	pca, adaptor := initPCA953xTestDriver()
	adaptor.i2cReadImpl = func([]byte) (int, error) {
		return 0, expectedReadError
	}
	// act
	_, err := pca.ReadGPIO(2) // index doesn't matter
	// assert
	gobottest.Assert(t, err, expectedReadError)
}

func TestPCA953xCalcPsc(t *testing.T) {
	// arrange
	var pca953xCalcPscTests = []pca953xCalcPscTest{
		{period: 0.0065, expectedVal: 0, expectedErr: ErrToSmallPeriod},
		{period: 0.0066, expectedVal: 0, expectedErr: nil},
		{period: 1, expectedVal: 151, expectedErr: nil},
		{period: 1.684, expectedVal: 255, expectedErr: nil},
		{period: 1.685, expectedVal: 255, expectedErr: ErrToBigPeriod},
	}
	for _, tp := range pca953xCalcPscTests {
		// act
		val, err := pca953xCalcPsc(tp.period)
		// assert
		gobottest.Assert(t, err, tp.expectedErr)
		gobottest.Assert(t, val, tp.expectedVal)
	}
}

func TestPCA953xCalcPeriod(t *testing.T) {
	// arrange
	var pca953xCalcPeriodTests = []pca953xCalcPeriodTest{
		{psc: 0, expectedVal: 0.0066},
		{psc: 1, expectedVal: 0.0132},
		{psc: 151, expectedVal: 1},
		{psc: 255, expectedVal: 1.6842},
	}
	for _, tp := range pca953xCalcPeriodTests {
		// act
		val := pca953xCalcPeriod(tp.psc)
		// assert
		gobottest.Assert(t, float32(math.Round(float64(val)*10000)/10000), tp.expectedVal)
	}
}

func TestPCA953xCalcPwm(t *testing.T) {
	// arrange
	var pca953xCalcPwmTests = []pca953xCalcPwmTest{
		{percent: -0.1, expectedVal: 0, expectedErr: ErrToSmallDutyCycle},
		{percent: 0, expectedVal: 0, expectedErr: nil},
		{percent: 49.9, expectedVal: 127, expectedErr: nil},
		{percent: 50, expectedVal: 128, expectedErr: nil},
		{percent: 100, expectedVal: 255, expectedErr: nil},
		{percent: 100.1, expectedVal: 255, expectedErr: ErrToBigDutyCycle},
	}
	for _, tp := range pca953xCalcPwmTests {
		// act
		val, err := pca953xCalcPwm(tp.percent)
		// assert
		gobottest.Assert(t, err, tp.expectedErr)
		gobottest.Assert(t, val, tp.expectedVal)
	}
}

func TestPCA953xCalcDutyCyclePercent(t *testing.T) {
	// arrange
	var pca953xCalcPwmTests = []pca953xCalcDutyCycleTest{
		{pwm: 0, expectedVal: 0},
		{pwm: 127, expectedVal: 49.8},
		{pwm: 128, expectedVal: 50.2},
		{pwm: 255, expectedVal: 100},
	}
	for _, tp := range pca953xCalcPwmTests {
		// act
		val := pca953xCalcDutyCyclePercent(tp.pwm)
		// assert
		gobottest.Assert(t, float32(math.Round(float64(val)*10)/10), tp.expectedVal)
	}
}

func TestPCA953xReadRegister(t *testing.T) {
	// arrange
	const expectedRegAddress = PCA953xRegister(0x03)
	const expectedReadByteCount = 1
	const expectedRegVal = uint8(0x04)
	readByteCount := 0
	pca, adaptor := initPCA953xTestDriver()
	// prepare all writes
	numCallsWrite := 0
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		numCallsWrite++
		return 0, nil
	}
	// prepare all reads
	numCallsRead := 0
	adaptor.i2cReadImpl = func(b []byte) (int, error) {
		numCallsRead++
		readByteCount = len(b)
		b[0] = expectedRegVal
		return readByteCount, nil
	}
	// act
	val, err := pca.readRegister(expectedRegAddress)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, numCallsRead, 1)
	gobottest.Assert(t, numCallsWrite, 1)
	gobottest.Assert(t, val, expectedRegVal)
	gobottest.Assert(t, readByteCount, expectedReadByteCount)
	gobottest.Assert(t, len(adaptor.written), 1)
	gobottest.Assert(t, adaptor.written[0], uint8(expectedRegAddress))
}

func TestPCA953xWriteRegister(t *testing.T) {
	// arrange
	const expectedRegAddress = PCA953xRegister(0x03)
	const expectedRegVal = uint8(0x97)
	const expectedByteCount = 2
	pca, adaptor := initPCA953xTestDriver()
	// prepare all writes
	numCallsWrite := 0
	adaptor.i2cWriteImpl = func(b []byte) (int, error) {
		numCallsWrite++
		return 0, nil
	}
	// act
	err := pca.writeRegister(expectedRegAddress, expectedRegVal)
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, numCallsWrite, 1)
	gobottest.Assert(t, numCallsWrite, 1)
	gobottest.Assert(t, len(adaptor.written), expectedByteCount)
	gobottest.Assert(t, adaptor.written[0], uint8(expectedRegAddress))
	gobottest.Assert(t, adaptor.written[1], uint8(expectedRegVal))
}
