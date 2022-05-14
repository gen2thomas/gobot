package gpio

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"gobot.io/x/gobot"
)

const (
	HD44780_CLEARDISPLAY        = 0x01
	HD44780_RETURNHOME          = 0x02
	HD44780_ENTRYMODESET        = 0x04
	HD44780_DISPLAYCONTROL      = 0x08
	HD44780_CURSORSHIFT         = 0x10
	HD44780_FUNCTIONSET         = 0x20
	HD44780_SETCGRAMADDR        = 0x40
	HD44780_SETDDRAMADDR        = 0x80
	HD44780_ENTRYRIGHT          = 0x00
	HD44780_ENTRYLEFT           = 0x02
	HD44780_ENTRYSHIFTINCREMENT = 0x01
	HD44780_ENTRYSHIFTDECREMENT = 0x00
	HD44780_DISPLAYON           = 0x04
	HD44780_DISPLAYOFF          = 0x00
	HD44780_CURSORON            = 0x02
	HD44780_CURSOROFF           = 0x00
	HD44780_BLINKON             = 0x01
	HD44780_BLINKOFF            = 0x00
	HD44780_DISPLAYMOVE         = 0x08
	HD44780_CURSORMOVE          = 0x00
	HD44780_MOVERIGHT           = 0x04
	HD44780_MOVELEFT            = 0x00
	HD44780_1LINE               = 0x00
	HD44780_2LINE               = 0x08
	HD44780_5x8DOTS             = 0x00
	HD44780_5x10DOTS            = 0x04
	HD44780_4BITBUS             = 0x00
	HD44780_8BITBUS             = 0x10
)

const (
	HD44780_2NDLINEOFFSET = 0x40
)

// data bus mode
type HD44780BusMode int

const (
	HD44780_4BITMODE HD44780BusMode = iota + 1
	HD44780_8BITMODE
)

// databit pins
type HD44780DataPin struct {
	D0 string // not used if 4bit mode
	D1 string // not used if 4bit mode
	D2 string // not used if 4bit mode
	D3 string // not used if 4bit mode
	D4 string
	D5 string
	D6 string
	D7 string
}

// HD44780Driver is the gobot driver for the HD44780 LCD controller
// Datasheet: https://www.sparkfun.com/datasheets/LCD/HD44780.pdf
type HD44780Driver struct {
	name          string
	cols          int
	rows          int
	rowOffsets    [4]int
	busMode       HD44780BusMode
	pinRS         *DirectPinDriver
	pinEN         *DirectPinDriver
	pinRW         *DirectPinDriver
	pinDataBits   []*DirectPinDriver
	displayCtrl   int
	displayFunc   int
	displayMode   int
	checkBusyFlag bool
	connection    gobot.Connection
	gobot.Commander
	mutex *sync.Mutex // mutex is needed for sequences, like CreateChar(), Write(), Start()
}

// NewHD44780Driver return a new HD44780Driver
// a: gobot.Conenction
// cols: lcd columns
// rows: lcd rows
// busMode: 4Bit or 8Bit
// pinRS: register select pin
// pinEN: clock enable pin
// pinDataBits: databit pins
func NewHD44780Driver(a gobot.Connection, cols int, rows int, busMode HD44780BusMode, pinRS string, pinEN string, pinDataBits HD44780DataPin) *HD44780Driver {
	h := &HD44780Driver{
		name:       "HD44780Driver",
		cols:       cols,
		rows:       rows,
		busMode:    busMode,
		pinRS:      NewDirectPinDriver(a, pinRS),
		pinEN:      NewDirectPinDriver(a, pinEN),
		connection: a,
		Commander:  gobot.NewCommander(),
		mutex:      &sync.Mutex{},
	}

	if h.busMode == HD44780_4BITMODE {
		h.pinDataBits = make([]*DirectPinDriver, 4)
		h.pinDataBits[0] = NewDirectPinDriver(a, pinDataBits.D4)
		h.pinDataBits[1] = NewDirectPinDriver(a, pinDataBits.D5)
		h.pinDataBits[2] = NewDirectPinDriver(a, pinDataBits.D6)
		h.pinDataBits[3] = NewDirectPinDriver(a, pinDataBits.D7)
	} else {
		h.pinDataBits = make([]*DirectPinDriver, 8)
		h.pinDataBits[0] = NewDirectPinDriver(a, pinDataBits.D0)
		h.pinDataBits[1] = NewDirectPinDriver(a, pinDataBits.D1)
		h.pinDataBits[2] = NewDirectPinDriver(a, pinDataBits.D2)
		h.pinDataBits[3] = NewDirectPinDriver(a, pinDataBits.D3)
		h.pinDataBits[4] = NewDirectPinDriver(a, pinDataBits.D4)
		h.pinDataBits[5] = NewDirectPinDriver(a, pinDataBits.D5)
		h.pinDataBits[6] = NewDirectPinDriver(a, pinDataBits.D6)
		h.pinDataBits[7] = NewDirectPinDriver(a, pinDataBits.D7)
	}

	h.rowOffsets[0] = 0x00
	h.rowOffsets[1] = HD44780_2NDLINEOFFSET
	h.rowOffsets[2] = 0x00 + cols
	h.rowOffsets[3] = HD44780_2NDLINEOFFSET + cols

	/* TODO : Add commands */

	return h
}

// SetRWPin initializes the RW pin
func (h *HD44780Driver) SetRWPin(pinRW string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	log.Println("SetRWPin called")
	h.pinRW = NewDirectPinDriver(h.connection, pinRW)
}

// SetBusyFlagCheck to the given state
func (h *HD44780Driver) SetBusyFlagCheck(state bool) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if state && h.pinRW == nil {
		return fmt.Errorf("The pinRW must be set for busy flag check. Please use SetRWPin() before.")
	}
	h.checkBusyFlag = state
	return nil
}

// Halt implements the Driver interface
func (h *HD44780Driver) Halt() error { return nil }

// Name returns the HD44780Driver name
func (h *HD44780Driver) Name() string { return h.name }

// SetName sets the HD44780Driver name
func (h *HD44780Driver) SetName(n string) { h.name = n }

// Connecton returns the HD44780Driver Connection
func (h *HD44780Driver) Connection() gobot.Connection {
	return h.connection
}

// Start initializes the HD44780 LCD controller
// refer to page 45/46 of hitachi HD44780 datasheet
func (h *HD44780Driver) Start() (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, bitPin := range h.pinDataBits {
		if bitPin.Pin() == "" {
			return errors.New("Initialization error")
		}
	}

	time.Sleep(50 * time.Millisecond)

	if err := h.activateWriteMode(); err != nil {
		return err
	}

	if h.busMode == HD44780_4BITMODE {
		if err := h.writeDataPins(0x03); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)

		if err := h.writeDataPins(0x03); err != nil {
			return err
		}
		time.Sleep(100 * time.Microsecond)

		if err := h.writeDataPins(0x03); err != nil {
			return err
		}
		time.Sleep(100 * time.Microsecond)

		if err := h.writeDataPins(0x02); err != nil {
			return err
		}
	} else {
		if err := h.sendCommand(0x30, "start1"); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)

		if err := h.sendCommand(0x30, "start2"); err != nil {
			return err
		}
		time.Sleep(100 * time.Microsecond)

		if err := h.sendCommand(0x30, "start3"); err != nil {
			return err
		}
	}
	time.Sleep(100 * time.Microsecond)

	if h.busMode == HD44780_4BITMODE {
		h.displayFunc |= HD44780_4BITBUS
	} else {
		h.displayFunc |= HD44780_8BITBUS
	}

	if h.rows > 1 {
		h.displayFunc |= HD44780_2LINE
	} else {
		h.displayFunc |= HD44780_1LINE
	}

	h.displayFunc |= HD44780_5x8DOTS
	h.displayCtrl = HD44780_DISPLAYON | HD44780_BLINKOFF | HD44780_CURSOROFF
	h.displayMode = HD44780_ENTRYLEFT | HD44780_ENTRYSHIFTDECREMENT

	time.Sleep(1 * time.Millisecond)
	if err := h.sendCommand(HD44780_DISPLAYCONTROL|h.displayCtrl, "start4"); err != nil {
		return err
	}
	time.Sleep(5 * time.Millisecond)
	if err := h.sendCommand(HD44780_FUNCTIONSET|h.displayFunc, "start5"); err != nil {
		return err
	}
	if err := h.sendCommand(HD44780_ENTRYMODESET|h.displayMode, "start6"); err != nil {
		return err
	}

	return h.clear()
}

// Write output text to the display
func (h *HD44780Driver) Write(message string) (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	col := 0
	if (h.displayMode & HD44780_ENTRYLEFT) == 0 {
		col = h.cols - 1
	}

	row := 0
	for _, c := range message {
		if c == '\n' {
			row++
			if err := h.setCursor(col, row, "Write-setCursor"); err != nil {
				return err
			}
			continue
		}
		if err := h.writeChar(int(c), "Write-writeChar"); err != nil {
			return err
		}
	}

	return nil
}

// Clear clear the display
func (h *HD44780Driver) Clear() (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.clear()
}

// Home return cursor to home
func (h *HD44780Driver) Home() (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if err := h.sendCommand(HD44780_RETURNHOME, "Home"); err != nil {
		return err
	}
	time.Sleep(2 * time.Millisecond)

	return nil
}

// SetCursor move the cursor to the specified position
func (h *HD44780Driver) SetCursor(col int, row int) (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.setCursor(col, row, "SetCursor")
}

// Display turn the display on and off
func (h *HD44780Driver) Display(on bool) (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if on {
		h.displayCtrl |= HD44780_DISPLAYON
	} else {
		h.displayCtrl &= ^HD44780_DISPLAYON
	}

	return h.sendCommand(HD44780_DISPLAYCONTROL|h.displayCtrl, "Display on/off")
}

// Cursor turn the cursor on and off
func (h *HD44780Driver) Cursor(on bool) (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if on {
		h.displayCtrl |= HD44780_CURSORON
	} else {
		h.displayCtrl &= ^HD44780_CURSORON
	}

	return h.sendCommand(HD44780_DISPLAYCONTROL|h.displayCtrl, "Cursor on/off")
}

// Blink turn the blink on and off
func (h *HD44780Driver) Blink(on bool) (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if on {
		h.displayCtrl |= HD44780_BLINKON
	} else {
		h.displayCtrl &= ^HD44780_BLINKON
	}

	return h.sendCommand(HD44780_DISPLAYCONTROL|h.displayCtrl, "Blink on/off")
}

// ScrollLeft scroll text left
func (h *HD44780Driver) ScrollLeft() (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.sendCommand(HD44780_CURSORSHIFT|HD44780_DISPLAYMOVE|HD44780_MOVELEFT, "Scroll left")
}

// ScrollRight scroll text right
func (h *HD44780Driver) ScrollRight() (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.sendCommand(HD44780_CURSORSHIFT|HD44780_DISPLAYMOVE|HD44780_MOVERIGHT, "Scroll right")
}

// LeftToRight display text from left to right
func (h *HD44780Driver) LeftToRight() (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.displayMode |= HD44780_ENTRYLEFT
	return h.sendCommand(HD44780_ENTRYMODESET|h.displayMode, "Left to right")
}

// RightToLeft display text from right to left
func (h *HD44780Driver) RightToLeft() (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.displayMode &= ^HD44780_ENTRYLEFT
	return h.sendCommand(HD44780_ENTRYMODESET|h.displayMode, "Right to left")
}

// SendCommand send control command
func (h *HD44780Driver) SendCommand(data int, sender string) (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.sendCommand(data, sender)
}

// WriteChar output a character to the display
func (h *HD44780Driver) WriteChar(data int) (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.writeChar(data, "WriteChar")
}

// CreateChar create custom character
func (h *HD44780Driver) CreateChar(pos int, charMap [8]byte) (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if pos > 7 {
		return errors.New("can't set a custom character at a position greater than 7")
	}

	if err := h.sendCommand(HD44780_SETCGRAMADDR|(pos<<3), "CreateChar-sendCommand"); err != nil {
		return err
	}

	for i := range charMap {
		if err := h.writeChar(int(charMap[i]), "CreateChar-writeChar"); err != nil {
			return err
		}
	}

	return nil
}

func (h *HD44780Driver) ReadAcAndDr() (ac int, dr int, err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// the read value represents the busy flag and address counter
	// BF  AC6 AC5 AC4 | AC3 AC2 AC1 AC0
	//
	// followed by the data register
	// DR7 DR6 DR5 DR4 | DR3 DR2 DR1 DR0
	//
	// no need to change RS, RW after that, because
	// this must be done by next write or read

	if ac, err = h.waitForBusyFlagAndReadAddressCounter(false, "Read AC & DR"); err != nil {
		return
	}

	if dr, err = h.readDataRegister(); err != nil {
		return
	}

	log.Printf("AC: 0x%X, DR: 0x%X", ac, dr)

	return
}

func (h *HD44780Driver) sendCommand(data int, sender string) (err error) {
	if err := h.waitForBusyFlag(sender); err != nil {
		return err
	}
	if err := h.activateWriteMode(); err != nil {
		return err
	}
	if err := h.pinRS.Off(); err != nil {
		return err
	}
	if h.busMode == HD44780_4BITMODE {
		if err := h.writeDataPins(data >> 4); err != nil {
			return err
		}
	}

	return h.writeDataPins(data)
}

func (h *HD44780Driver) writeChar(data int, sender string) (err error) {

	if err := h.waitForBusyFlag(sender); err != nil {
		return err
	}

	if err := h.activateWriteMode(); err != nil {
		return err
	}

	if err := h.pinRS.On(); err != nil {
		return err
	}
	if h.busMode == HD44780_4BITMODE {
		if err := h.writeDataPins(data >> 4); err != nil {
			return err
		}
	}

	return h.writeDataPins(data)
}

func (h *HD44780Driver) clear() (err error) {
	if err := h.sendCommand(HD44780_CLEARDISPLAY, "Clear"); err != nil {
		return err
	}
	time.Sleep(2 * time.Millisecond)

	return nil
}

func (h *HD44780Driver) setCursor(col int, row int, sender string) (err error) {
	if col < 0 || row < 0 || col >= h.cols || row >= h.rows {
		return fmt.Errorf("Invalid position value (%d, %d), range (%d, %d)", col, row, h.cols-1, h.rows-1)
	}

	return h.sendCommand(HD44780_SETDDRAMADDR|col+h.rowOffsets[row], sender)
}

func (h *HD44780Driver) waitForBusyFlag(sender string) (err error) {
	_, err = h.waitForBusyFlagAndReadAddressCounter(true, sender)
	return
}

func (h *HD44780Driver) waitForBusyFlagAndReadAddressCounter(onlyBF bool, sender string) (result int, err error) {
	// the busy flag is the highest bit (D7) in read mode
	//
	// sequence:
	// RS = "0", RW = "1" (means read)
	// EN = "0", wait, EN = "1"
	// read pin --> wait, read again, until = 0
	// read AC or EN = "0", wait, EN = "1", wait, EN ="0" (to simulate AC-read)
	// store to AC
	//
	if h.pinRW == nil || !h.checkBusyFlag {
		return
	}
	// ensure disabled state
	if err = h.pinEN.Off(); err != nil {
		return
	}
	// activate instruction register
	if err = h.pinRS.Off(); err != nil {
		return
	}
	// activate read mode
	if err = h.pinRW.On(); err != nil {
		return
	}

	// start reading
	if err = h.pinEN.On(); err != nil {
		return
	}

	// repeat until BF is "0"
	// reset formerly set bf pin
	bfPin := 7
	if h.busMode == HD44780_4BITMODE {
		bfPin = 3
	}
	i := 1
	maxWait := 20
	var waitTime time.Duration
	for ; i <= maxWait; i++ {
		var val int
		if val, err = h.pinDataBits[bfPin].DigitalRead(); err != nil {
			return
		}
		if (val & 0x01) == 0 {
			break
		}
		log.Printf("busy flag is set %d/%d for %s", i, maxWait, sender)
		waitTime = time.Duration(i*i) * time.Millisecond
		time.Sleep(waitTime)
	}

	if !onlyBF {
		if result, err = h.readDataPins(); err != nil {
			return
		}
		if h.busMode == HD44780_4BITMODE {
			var data int
			if data, err = h.readDataPins(); err != nil {
				return
			}
			result = result<<4 | data
		}
	} else {
		if err = h.pinEN.Off(); err != nil {
			return
		}
		time.Sleep(1 * time.Microsecond)
		if h.busMode == HD44780_4BITMODE {
			// simulate reading of AC
			if err = h.fallingEdge(); err != nil {
				return
			}
		}
	}

	if i >= maxWait {
		return 0, fmt.Errorf("busy flag remains on after wait %s", waitTime)
	}
	return
}

func (h *HD44780Driver) readDataRegister() (result int, err error) {
	// sequence:
	// RW = "1", RS = "1", EN ="1"
	// read data pins, EN ="0"
	//
	// store to DR
	//
	// RW is still "1"
	if err = h.pinRS.On(); err != nil {
		return
	}
	if err = h.pinEN.On(); err != nil {
		return
	}
	if result, err = h.readDataPins(); err != nil {
		return
	}
	if h.busMode == HD44780_4BITMODE {
		var data int
		if data, err = h.readDataPins(); err != nil {
			return
		}
		result = result<<4 | data
	}

	return
}

func (h *HD44780Driver) writeDataPins(data int) (err error) {
	for i, pin := range h.pinDataBits {
		if ((data >> i) & 0x01) == 0x01 {
			if err := pin.On(); err != nil {
				return err
			}
		} else {
			if err := pin.Off(); err != nil {
				return err
			}
		}
	}
	return h.fallingEdge()
}

func (h *HD44780Driver) readDataPins() (data int, err error) {
	var val int
	for i, pin := range h.pinDataBits {
		if val, err = pin.DigitalRead(); err != nil {
			return
		}
		data = data | (val&0x1)<<i
	}
	return data, h.fallingEdge()
}

// fallingEdge creates falling edge to trigger data transmission
func (h *HD44780Driver) fallingEdge() (err error) {
	if err := h.pinEN.On(); err != nil {
		return err
	}
	time.Sleep(1 * time.Microsecond)

	if err := h.pinEN.Off(); err != nil {
		return err
	}
	// fastest write operation at 190kHz mode takes 53 us
	time.Sleep(60 * time.Microsecond)

	return nil
}

func (h *HD44780Driver) activateWriteMode() (err error) {
	if h.pinRW == nil {
		return
	}
	return h.pinRW.Off()
}
