package i2c

import (
	"strings"
	"testing"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/gobottest"
)

var _ gobot.Driver = (*VL53L0XDriver)(nil)

func initTestVL53L0XDriver() (driver *VL53L0XDriver) {
	driver, _ = initTestVL53L0XDriverWithStubbedAdaptor()
	return
}

func initTestVL53L0XDriverWithStubbedAdaptor() (*VL53L0XDriver, *i2cTestAdaptor) {
	a := newI2cTestAdaptor()
	d := NewVL53L0XDriver(a)
	if err := d.Start(); err != nil {
		panic(err)
	}
	return d, a
}

func TestNewVL53L0XDriver(t *testing.T) {
	var di interface{} = NewVL53L0XDriver(newI2cTestAdaptor())
	d, ok := di.(*VL53L0XDriver)
	if !ok {
		t.Errorf("NewVL53L0XDriver() should have returned a *VL53L0XDriver")
	}
	gobottest.Refute(t, d.Driver, nil)
	gobottest.Assert(t, strings.HasPrefix(d.name, "VL53L0X"), true)
}

func TestVL53L0XDriverOptions(t *testing.T) {
	// This is a general test, that options are applied in constructor by using the common WithBus() option.
	// Further tests for options can also be done by call of "WithOption(val)(d)".
	l := NewVL53L0XDriver(newI2cTestAdaptor(), WithBus(2))
	gobottest.Assert(t, l.GetBusOrDefault(1), 2)
}

func TestVL53L0XDistance(t *testing.T) {
	// sequence to read the distance:
	// * read control register for get current state and ensure an clock mode is set
	// * write the control register (stop counting)
	// * create the values for date registers (default is 24h mode)
	// * write the clock and calendar registers with auto increment
	// * write the control register (start counting)
	// arrange
	d, a := initTestVL53L0XDriverWithStubbedAdaptor()
	a.i2cReadImpl = func(b []byte) (int, error) {
		return len(b), nil
	}
	// act
	distance, err := d.Distance()

	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, distance, int(10))
}
