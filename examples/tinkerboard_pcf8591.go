// +build example
//
// Do not build by default.

package main

import (
	"fmt"
	"log"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/tinkerboard"
)

func main() {
	// This driver was tested with Tinkerboard and this board with temperature & brightness sensor:
	// https://www.makershop.de/download/YL_40_PCF8591.pdf
	//
	// Wiring
	// PWR  Tinkerboard: 1 (+3.3V, VCC), 6, 9, 14, 20 (GND)
	// I2C1 Tinkerboard: 3 (SDA), 5 (SCL)
	// PCF8591 plate: wire AOUT --> AIN2 for this example
	board := tinkerboard.NewAdaptor()
	pcf := i2c.NewPCF8591Driver(board, i2c.WithBus(1),
		i2c.WithPCF8591RescaleInput(0, 1000, 0),
		i2c.WithPCF8591RescaleInput(1, 255, 0),
		i2c.WithPCF8591RescaleInput(3, 100, -100))
	var val int
	var err error

	// brightness sensor, high brightness - low raw value, scaled to 0..1000 (high brightness - high value)
	descLight := "s.0"
	// temperature sensor, high temperature - low raw value, scaled to 0..255 (high temperature - high value)
	// sometimes buggy, because not properly grounded
	descTemp := "s.1"
	// wired to AOUT, scaled to voltage 3300mV (the default)
	descAIN2 := "s.2"
	// adjustable resistor, turn clockwise will lower the raw value, scaled to -100..+100% (clockwise)
	descResi := "s.3"
	// the LED light is visible above ~1.6V
	writeVal := 1500

	work := func() {
		gobot.Every(1000*time.Millisecond, func() {
			if err := pcf.AnalogWrite(writeVal); err != nil {
				fmt.Println(err)
			} else {
				log.Printf(" %d mV written", writeVal)
				writeVal = writeVal + 100
				if writeVal > 3300 {
					writeVal = 0
				}
			}

			if val, err = pcf.AnalogRead(descLight); err != nil {
				fmt.Println(err)
			} else {
				log.Printf("Brightness (%s): %d [0..1000]", descLight, val)
			}

			if val, err = pcf.AnalogRead(descTemp); err != nil {
				fmt.Println(err)
			} else {
				log.Printf("Temperature (%s): %d [0..255]", descTemp, val)
			}

			if val, err = pcf.AnalogRead(descAIN2); err != nil {
				fmt.Println(err)
			} else {
				log.Printf("Read AOUT (%s): %d mV [0..3300]", descAIN2, val)
			}

			if val, err = pcf.AnalogRead(descResi); err != nil {
				fmt.Println(err)
			} else {
				log.Printf("Resistor (%s): %d %% [-100..+100]", descResi, val)
			}
		})
	}

	robot := gobot.NewRobot("pcfBot",
		[]gobot.Connection{board},
		[]gobot.Device{pcf},
		work,
	)

	robot.Start()
}
