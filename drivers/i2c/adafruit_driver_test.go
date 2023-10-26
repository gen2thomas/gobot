package i2c

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gobot.io/x/gobot/v2"
)

// this ensures that the implementation implements the gobot.Driver interface
var _ gobot.Driver = (*AdafruitMotorHatDriver)(nil)

// --------- HELPERS
func initTestAdafruitMotorHatDriver() (driver *AdafruitMotorHatDriver) {
	driver, _ = initTestAdafruitMotorHatDriverWithStubbedAdaptor()
	return
}

func initTestAdafruitMotorHatDriverWithStubbedAdaptor() (*AdafruitMotorHatDriver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor()
	return NewAdafruitMotorHatDriver(adaptor), adaptor
}

// --------- TESTS
func TestNewAdafruitMotorHatDriver(t *testing.T) {
	var di interface{} = NewAdafruitMotorHatDriver(newI2cTestAdaptor())
	d, ok := di.(*AdafruitMotorHatDriver)
	if !ok {
		t.Errorf("AdafruitMotorHatDriver() should have returned a *AdafruitMotorHatDriver")
	}
	assert.True(t, strings.HasPrefix(d.Name(), "AdafruitMotorHat"))
}

// Methods
func TestAdafruitMotorHatDriverStart(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()
	assert.NotNil(t, ada.Connection())
	assert.NoError(t, ada.Start())
}

func TestAdafruitMotorHatDriverStartWriteError(t *testing.T) {
	d, adaptor := initTestAdafruitMotorHatDriverWithStubbedAdaptor()
	adaptor.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}
	assert.ErrorContains(t, d.Start(), "write error")
}

func TestAdafruitMotorHatDriverStartReadError(t *testing.T) {
	d, adaptor := initTestAdafruitMotorHatDriverWithStubbedAdaptor()
	adaptor.i2cReadImpl = func([]byte) (int, error) {
		return 0, errors.New("read error")
	}
	assert.ErrorContains(t, d.Start(), "read error")
}

func TestAdafruitMotorHatDriverStartConnectError(t *testing.T) {
	d, adaptor := initTestAdafruitMotorHatDriverWithStubbedAdaptor()
	adaptor.Testi2cConnectErr(true)
	assert.ErrorContains(t, d.Start(), "Invalid i2c connection")
}

func TestAdafruitMotorHatDriverHalt(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Halt())
}

func TestSetHatAddresses(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	motorHatAddr := 0x61
	servoHatAddr := 0x41
	assert.NoError(t, ada.SetMotorHatAddress(motorHatAddr))
	assert.NoError(t, ada.SetServoHatAddress(servoHatAddr))
}

func TestAdafruitMotorHatDriverSetServoMotorFreq(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())

	freq := 60.0
	err := ada.SetServoMotorFreq(freq)
	assert.NoError(t, err)
}

func TestAdafruitMotorHatDriverSetServoMotorFreqError(t *testing.T) {
	ada, a := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())
	a.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}

	freq := 60.0
	assert.ErrorContains(t, ada.SetServoMotorFreq(freq), "write error")
}

func TestAdafruitMotorHatDriverSetServoMotorPulse(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())

	var channel byte = 7
	var on int32 = 1234
	var off int32 = 4321
	err := ada.SetServoMotorPulse(channel, on, off)
	assert.NoError(t, err)
}

func TestAdafruitMotorHatDriverSetServoMotorPulseError(t *testing.T) {
	ada, a := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())
	a.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}

	var channel byte = 7
	var on int32 = 1234
	var off int32 = 4321
	assert.ErrorContains(t, ada.SetServoMotorPulse(channel, on, off), "write error")
}

func TestAdafruitMotorHatDriverSetDCMotorSpeed(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())

	dcMotor := 1
	var speed int32 = 255
	err := ada.SetDCMotorSpeed(dcMotor, speed)
	assert.NoError(t, err)
}

func TestAdafruitMotorHatDriverSetDCMotorSpeedError(t *testing.T) {
	ada, a := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())
	a.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}

	assert.ErrorContains(t, ada.SetDCMotorSpeed(1, 255), "write error")
}

func TestAdafruitMotorHatDriverRunDCMotor(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())

	dcMotor := 1
	assert.NoError(t, ada.RunDCMotor(dcMotor, AdafruitForward))
	assert.NoError(t, ada.RunDCMotor(dcMotor, AdafruitBackward))
	assert.NoError(t, ada.RunDCMotor(dcMotor, AdafruitRelease))
}

func TestAdafruitMotorHatDriverRunDCMotorError(t *testing.T) {
	ada, a := initTestAdafruitMotorHatDriverWithStubbedAdaptor()
	assert.NoError(t, ada.Start())
	a.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}

	dcMotor := 1
	assert.ErrorContains(t, ada.RunDCMotor(dcMotor, AdafruitForward), "write error")
	assert.ErrorContains(t, ada.RunDCMotor(dcMotor, AdafruitBackward), "write error")
	assert.ErrorContains(t, ada.RunDCMotor(dcMotor, AdafruitRelease), "write error")
}

func TestAdafruitMotorHatDriverSetStepperMotorSpeed(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())

	stepperMotor := 1
	rpm := 30
	assert.NoError(t, ada.SetStepperMotorSpeed(stepperMotor, rpm))
}

func TestAdafruitMotorHatDriverStepperMicroStep(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())

	// NOTE: not using the direction and style constants to prevent importing
	// the i2c package
	stepperMotor := 0
	steps := 50
	err := ada.Step(stepperMotor, steps, 1, 3)
	assert.NoError(t, err)
}

func TestAdafruitMotorHatDriverStepperSingleStep(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())

	// NOTE: not using the direction and style constants to prevent importing
	// the i2c package
	stepperMotor := 0
	steps := 50
	err := ada.Step(stepperMotor, steps, 1, 0)
	assert.NoError(t, err)
}

func TestAdafruitMotorHatDriverStepperDoubleStep(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())

	// NOTE: not using the direction and style constants to prevent importing
	// the i2c package
	stepperMotor := 0
	steps := 50
	err := ada.Step(stepperMotor, steps, 1, 1)
	assert.NoError(t, err)
}

func TestAdafruitMotorHatDriverStepperInterleaveStep(t *testing.T) {
	ada, _ := initTestAdafruitMotorHatDriverWithStubbedAdaptor()

	assert.NoError(t, ada.Start())

	// NOTE: not using the direction and style constants to prevent importing
	// the i2c package
	stepperMotor := 0
	steps := 50
	err := ada.Step(stepperMotor, steps, 1, 2)
	assert.NoError(t, err)
}

func TestAdafruitMotorHatDriverSetName(t *testing.T) {
	d := initTestAdafruitMotorHatDriver()
	d.SetName("TESTME")
	assert.Equal(t, "TESTME", d.Name())
}

func TestAdafruitMotorHatDriverOptions(t *testing.T) {
	d := NewAdafruitMotorHatDriver(newI2cTestAdaptor(), WithBus(2))
	assert.Equal(t, 2, d.GetBusOrDefault(1))
}
