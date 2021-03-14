package mcp2221

import (
	"fmt"
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/sysfs"
)

type sysfsPin struct {
	pin    int
	pwmPin int
}

const mcp2221MinBus = 6
const mcp2221DefaultBus = 6

// Adaptor represents a Gobot Adaptor for the mcp2221
type Adaptor struct {
	name               string	
	i2cBuses           [10]i2c.I2cDevice
	mutex              *sync.Mutex
}

// NewAdaptor creates a mcp2221 Adaptor
func NewAdaptor() *Adaptor {
	c := &Adaptor{
		name:    gobot.DefaultName("mcp2221"),
		mutex:   &sync.Mutex{},
	}

	return c
}

// Name returns the name of the Adaptor
func (c *Adaptor) Name() string { return c.name }

// SetName sets the name of the Adaptor
func (c *Adaptor) SetName(n string) { c.name = n }

// Connect initializes the board
func (c *Adaptor) Connect() (err error) {
	return nil
}

// Finalize closes connection to board and pins
func (c *Adaptor) Finalize() (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, bus := range c.i2cBuses {
		if bus != nil {
			if e := bus.Close(); e != nil {
				err = multierror.Append(err, e)
			}
		}
	}

	return
}

// GetConnection returns a connection to a device on a specified bus.
// Valid bus number is [mcp2221MinBus..9] which corresponds to /dev/i2c-* through /dev/i2c-9.
func (c *Adaptor) GetConnection(address int, bus int) (connection i2c.Connection, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if (bus < mcp2221MinBus) || (bus > 9) {
		return nil, fmt.Errorf("Bus number %d out of range", bus)
	}
	if c.i2cBuses[bus] == nil {
		c.i2cBuses[bus], err = sysfs.NewI2cDevice(fmt.Sprintf("/dev/i2c-%d", bus))
	}
	return i2c.NewConnection(c.i2cBuses[bus], address), err
}

// GetDefaultBus returns the default i2c bus for your platform
func (c *Adaptor) GetDefaultBus() int {
	return mcp2221DefaultBus
}
