// +build example
//
// Do not build by default.

package main

import (
	"fmt"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/digispark"
)

func main() {
	board := digispark.NewAdaptor()
	gpio := i2c.NewPCA9501Driver(board)
	var pin uint8 = 0
	var pinState uint8 = 0

	work := func() {
		gobot.Every(100*time.Millisecond, func() {
			fmt.Println("set Pin:", pin, "to:", pinState)
			gpio.WriteGPIO(pin, pinState)
			pin = pin + 1
			if pin >= 8 {
				pin = 0
				if pinState == 0 {
					pinState = 1
				} else {
					pinState = 0
				}
			}
		})
	}

	robot := gobot.NewRobot("rotatePinsI2c",
		[]gobot.Connection{board},
		[]gobot.Device{gpio},
		work,
	)

	err := robot.Start()
	if err != nil {
		fmt.Println(err)
	}
}
