package i2c

import "fmt"

const vl53L0XDefaultAddress = 0x52

const (
	vl53L0XRegRef1 = 0xC0 // start reference register 1..3 (0xC0, 0xC1, 0xC2)
	vl53L0XRegRef2 = 0x51 // start reference register 2 (2 bytes)
	vl53L0XRegRef3 = 0x61 // start reference register 3 (2 bytes)

	vl53L0XRef1Val1 = 0xEE   // content of reference register 1.1 after fresh reset
	vl53L0XRef1Val2 = 0xAA   // content of reference register 1.2 after fresh reset
	vl53L0XRef1Val3 = 0x10   // content of reference register 1.3 after fresh reset
	vl53L0XRef2Val  = 0x0099 // content of reference register 2 after fresh reset
	vl53L0XRef3Val  = 0x0000 // content of reference register 3 after fresh reset
)

// VL53L0XDriver is the Gobot driver for the Time-of-Flight ranging and gesture detection sensor.
//
// Datasheet: https://www.st.com/resource/en/datasheet/vl53l0x.pdf
// Important note: ST will not provide a register list, for further details please read:
// https://community.st.com/s/question/0D50X00009XkeHcSAJ/vl53l0x-register-map
//
// Reference implementations:
//  * https://github.com/adafruit/Adafruit_VL53L0X
//  * https://github.com/pololu/vl53l0x-arduino
type VL53L0XDriver struct {
	*Driver
}

// NewVL53L0XDriver creates a new driver for VL53L0X i2c device.
//
// Params:
//		conn Connector - the Adaptor to use with this Driver
//
// Optional params:
//		i2c.WithBus(int):	bus to use with this driver
//		i2c.WithAddress(int):	address to use with this driver
//
func NewVL53L0XDriver(a Connector, options ...func(Config)) *VL53L0XDriver {
	d := &VL53L0XDriver{
		Driver: NewDriver(a, "VL53L0X", vl53L0XDefaultAddress),
	}
	d.afterStart = d.initialize

	for _, option := range options {
		option(d)
	}

	// TODO: add commands to API
	return d
}

// Distance returns the current distance in cm
func (d *VL53L0XDriver) Distance() (int, error) {
	return 10, nil
}

func (d *VL53L0XDriver) initialize() error {
	data := make([]byte, 3)
	if err := d.connection.ReadBlockData(vl53L0XRegRef1, data); err != nil {
		return err
	}
	if data[0] != vl53L0XRef1Val1 {
		return fmt.Errorf("reference value (%d) should be %d", data[0], vl53L0XRef1Val1)
	}
	if data[1] != vl53L0XRef1Val2 {
		return fmt.Errorf("reference value (%d) should be %d", data[0], vl53L0XRef1Val1)
	}
	if data[2] != vl53L0XRef1Val3 {
		return fmt.Errorf("reference value (%d) should be %d", data[0], vl53L0XRef1Val1)
	}

	// TODO stopped here with implementing, because the initialization routine needs already some amount of time

	return nil
}
