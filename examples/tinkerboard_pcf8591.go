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
	pcf := i2c.NewPCF8591Driver(board, i2c.WithBus(1))
	var val int
	var err error

	// brightness sensor, high brightness - low value
	descLight := "s.0"
	// temperature sensor, high temperature - low value
	// sometimes buggy, because not properly grounded
	descTemp := "s.1"
	// wired to AOUT
	descAIN2 := "s.2"
	// adjustable resistor, turn clockwise will lower the value
	descResi := "s.3"
	// the LED light is visible above ~100
	writeVal := uint8(0)

	work := func() {
		gobot.Every(1000*time.Millisecond, func() {
			if err := pcf.AnalogWrite(writeVal); err != nil {
				fmt.Println(err)
			} else {
				log.Printf("Written: %d", writeVal)
				writeVal = writeVal + 10
			}

			if val, err = pcf.AnalogRead(descLight); err != nil {
				fmt.Println(err)
			} else {
				log.Printf("Read %s: %d", descLight, val)
			}

			if val, err = pcf.AnalogRead(descTemp); err != nil {
				fmt.Println(err)
			} else {
				log.Printf("Read %s: %d", descTemp, val)
			}

			if val, err = pcf.AnalogRead(descAIN2); err != nil {
				fmt.Println(err)
			} else {
				log.Printf("Read %s: %d", descAIN2, val)
			}

			if val, err = pcf.AnalogRead(descResi); err != nil {
				fmt.Println(err)
			} else {
				log.Printf("Read %s: %d", descResi, val)
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
