# mcp2221

The mcp2221 is a USB 2.0 to I2C/UART Protocol Converter with GPIO. This adaptor makes a gobot Board from each host system, which has a USB connector. For I2C it works like a bridge to the hosts I2C bus.

For more information about the Converter, go to [https://www.microchip.com/wwwproducts/en/MCP2221A](https://www.microchip.com/wwwproducts/en/MCP2221A).

## Breakout boards

Simple board from [Microchip](https://www.microchip.com/developmenttools/ProductDetails/PartNO/ADM00559)
Board to use with breadboard from [Adafruit](https://www.adafruit.com/product/4471)

## How to Install

### Install Go and Gobot
```
go get -d -u gobot.io/x/gobot/...
```

## How to Use

```go
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
	pca := i2c.NewPCA9501Driver(board)
	var pin uint8 = 0
	var pinState uint8 = 0

	work := func() {
		gobot.Every(100*time.Millisecond, func() {
			fmt.Println("set Pin:", pin, "to:", pinState)
			pca.WriteGPIO(pin, pinState)
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
		[]gobot.Device{pca},
		work,
	)

	err := robot.Start()
	if err != nil {
		fmt.Println(err)
	}
}
```

## How to Connect

To install drivers and customize your environment please have a look at [MCP2221A breakout module site](https://www.microchip.com/developmenttools/ProductDetails/PartNO/ADM00559)

### Guide for Linux
1. Download and build [Linux driver](https://ww1.microchip.com/downloads/en/DeviceDoc/mcp2221_0_1.tar.gz) according included "ReadMe" file.

2. Ensure libusb and libudev is installed. Howto install will depends on your Linux distribution, e.g. for Gentoo use:
```bash
emerge --ask libudev libusb
```

3. Load modules (when not already there) and check current devices
```bash
sudo modprobe i2c-dev
sudo ls -la /dev/i2*
```
Note the highest device before connecting the board, e.g. "/dev/i2c-5".

4. Setup your udev rules to change permissions for the device
```bash
echo '#MCP2221A adafruit breakout (USB to I2C etc.)' >> /dev/udev/rules.d/99_mcp2221.rules
echo 'SUBSYSTEM=="usb", ATTRS{idVendor}=="04d8", ATTR{idProduct}=="00dd", MODE="0666"' >> /dev/udev/rules.d/99_mcp2221.rules
echo KERNEL=="i2c-6", GROUP="i2c", MODE="0666" >> /dev/udev/rules.d/99_mcp2221.rules
``` 
The last line refer to the device "i2c-6", because the highest device was "i2c-5" without connected board. Please adjust to your highest device number plus one.

5. Add your user to the i2c group:
```bash
sudo usermod -aG i2c <username>
```

Reboot your system or reload the udev rules to to take effect.

6. Check for your device
Connect your device to your machine with an USB cable. Execute the script from the microchip installation guide and check for the device is recognized:
```bash
sudo lsusb | grep MCP2221
sudo ./driver_load.sh
sudo ls -la /dev/i2c-6
```

The last line sould show something like this "crw-rw-rw- 1 root i2c 89, 6 14. Mär 17:59 /dev/i2c-6", because the highest device was "i2c-5" without connected board. This should be your highest device number plus one.

When you have added another device to your board it should be recognized when scanning the bus "6" (respective your highest bus plus one):

```bash
i2cdetect -y 6
```

Example output for an PCA9501 GPIO (address 0x04) & EEPROM (address 0x04|0x40 = 0x44)

     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
00:          -- 04 -- -- -- -- -- -- -- -- -- -- -- 
10: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- 
20: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- 
30: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- 
40: -- -- -- -- 44 -- -- -- -- -- -- -- -- -- -- -- 
50: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- 
60: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- 
70: -- -- -- -- -- -- -- -- 

**IMPORTANT NOTE REGARDING I2C:** 
Scan for I2C devices using the `i2cdetect -l` command line tool can cause the I2C subsystem to malfunction until you reboot your system. Some systems reboot immediatelly when using `i2cdetect -l`, so be careful.

### UART interface (/dev/ttyACMx)
When you are interested on working UART, have a look at [this document](https://ww1.microchip.com/downloads/en/DeviceDoc/MCP2200_MCP2221_CDC_Linux_Readme.txt)
