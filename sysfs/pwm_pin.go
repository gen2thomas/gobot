package sysfs

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strconv"
	"syscall"
	"time"
)

const pwmPinErrorPattern = "%s failed for pin %s with %v"

// PWMPin is the interface for sysfs PWM interactions
type PWMPinner interface {
	// Export exports the pin for use by the operating system
	Export() error
	// Unexport unexports the pin and releases the pin from the operating system
	Unexport() error
	// Enable enables/disables the PWM pin
	Enable(bool) (err error)
	// Polarity returns the polarity either normal or inverted
	Polarity() (polarity string, err error)
	// InvertPolarity sets the polarity to inverted if called with true
	InvertPolarity(invert bool) (err error)
	// Period returns the current PWM period for pin
	Period() (period uint32, err error)
	// SetPeriod sets the current PWM period for pin
	SetPeriod(period uint32) (err error)
	// DutyCycle returns the duty cycle for the pin
	DutyCycle() (duty uint32, err error)
	// SetDutyCycle writes the duty cycle to the pin
	SetDutyCycle(duty uint32) (err error)
}

// PWMPinnerProvider is the interface that an Adaptor should implement to allow
// clients to obtain access to any PWMPin's available on that board.
type PWMPinnerProvider interface {
	PWMPin(string) (PWMPinner, error)
}

type PWMPin struct {
	pin     string
	Path    string
	enabled bool
	write   func(path string, data []byte) (i int, err error)
	read    func(path string) ([]byte, error)
}

// NewPwmPin returns a new pwmPin
func NewPWMPin(pin int) *PWMPin {
	return &PWMPin{
		pin:     strconv.Itoa(pin),
		enabled: false,
		Path:    "/sys/class/pwm/pwmchip0",
		read:    readPwmFile,
		write:   writePwmFile}
}

// Export writes pin to pwm export path
func (p *PWMPin) Export() error {
	_, err := p.write(p.pwmExportPath(), []byte(p.pin))
	if err != nil {
		// If EBUSY then the pin has already been exported
		e, ok := err.(*os.PathError)
		if !ok || e.Err != syscall.EBUSY {
			return fmt.Errorf(pwmPinErrorPattern, "Export", p.pin, err)
		}
	}

	// Pause to avoid race condition in case there is any udev rule
	// that changes file permissions on newly exported PWMPin. This
	// is a common circumstance when running as a non-root user.
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Unexport writes pin to pwm unexport path
func (p *PWMPin) Unexport() (err error) {
	if _, err = p.write(p.pwmUnexportPath(), []byte(p.pin)); err != nil {
		err = fmt.Errorf(pwmPinErrorPattern, "Unexport", p.pin, err)
	}
	return
}

// Enable writes value to pwm enable path
func (p *PWMPin) Enable(enable bool) (err error) {
	if p.enabled != enable {
		p.enabled = enable
		enableVal := 0
		if enable {
			enableVal = 1
		}
		if _, err = p.write(p.pwmEnablePath(), []byte(fmt.Sprintf("%v", enableVal))); err != nil {
			err = fmt.Errorf(pwmPinErrorPattern, "set Enable", p.pin, err)
		}
	}
	return
}

// Polarity returns current polarity value
func (p *PWMPin) Polarity() (polarity string, err error) {
	buf, err := p.read(p.pwmPolarityPath())
	if err != nil {
		return polarity, fmt.Errorf(pwmPinErrorPattern, "Polarity", p.pin, err)
	}
	if len(buf) == 0 {
		return "", nil
	}

	return string(buf), nil
}

// InvertPolarity sets the polarity to "inverted" when 'true' is given, otherwise to "normal"
func (p *PWMPin) InvertPolarity(invert bool) (err error) {
	polarity := "normal"
	if invert {
		polarity = "inverted"
	}
	return p.SetPolarity(polarity)
}

// SetPolarity writes value to pwm polarity path
func (p *PWMPin) SetPolarity(value string) (err error) {
	if p.enabled {
		return fmt.Errorf("Cannot set PWM polarity when enabled")
	}
	if _, err = p.write(p.pwmPolarityPath(), []byte(value)); err != nil {
		err = fmt.Errorf(pwmPinErrorPattern, "SetPolarity", p.pin, err)
	}
	return
}

// Period reads from pwm period path and returns value in nanoseconds
func (p *PWMPin) Period() (period uint32, err error) {
	buf, err := p.read(p.pwmPeriodPath())
	if err != nil {
		return period, fmt.Errorf(pwmPinErrorPattern, "Period", p.pin, err)
	}
	if len(buf) == 0 {
		return 0, nil
	}

	v := bytes.TrimRight(buf, "\n")
	val, e := strconv.Atoi(string(v))
	return uint32(val), e
}

// SetPeriod sets pwm period in nanoseconds
func (p *PWMPin) SetPeriod(period uint32) (err error) {
	if _, err = p.write(p.pwmPeriodPath(), []byte(fmt.Sprintf("%v", period))); err != nil {
		err = fmt.Errorf(pwmPinErrorPattern, "SetPeriod", p.pin, err)
	}
	return
}

// DutyCycle reads from pwm duty cycle path and returns value in nanoseconds
func (p *PWMPin) DutyCycle() (duty uint32, err error) {
	buf, err := p.read(p.pwmDutyCyclePath())
	if err != nil {
		return duty, fmt.Errorf(pwmPinErrorPattern, "DutyCycle", p.pin, err)
	}

	v := bytes.TrimRight(buf, "\n")
	val, e := strconv.Atoi(string(v))
	return uint32(val), e
}

// SetDutyCycle writes value to pwm duty cycle path
// duty is in nanoseconds
func (p *PWMPin) SetDutyCycle(duty uint32) (err error) {
	if _, err = p.write(p.pwmDutyCyclePath(), []byte(fmt.Sprintf("%v", duty))); err != nil {
		err = fmt.Errorf(pwmPinErrorPattern, "SetDutyCycle", p.pin, err)
	}
	return
}

// pwmExportPath returns export path
func (p *PWMPin) pwmExportPath() string {
	return path.Join(p.Path, "export")
}

// pwmUnexportPath returns unexport path
func (p *PWMPin) pwmUnexportPath() string {
	return path.Join(p.Path, "unexport")
}

// pwmDutyCyclePath returns duty_cycle path for specified pin
func (p *PWMPin) pwmDutyCyclePath() string {
	return path.Join(p.Path, "pwm"+p.pin, "duty_cycle")
}

// pwmPeriodPath returns period path for specified pin
func (p *PWMPin) pwmPeriodPath() string {
	return path.Join(p.Path, "pwm"+p.pin, "period")
}

// pwmEnablePath returns enable path for specified pin
func (p *PWMPin) pwmEnablePath() string {
	return path.Join(p.Path, "pwm"+p.pin, "enable")
}

// pwmPolarityPath returns polarity path for specified pin
func (p *PWMPin) pwmPolarityPath() string {
	return path.Join(p.Path, "pwm"+p.pin, "polarity")
}

func writePwmFile(path string, data []byte) (i int, err error) {
	file, err := OpenFile(path, os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		return
	}

	return file.Write(data)
}

func readPwmFile(path string) ([]byte, error) {
	file, err := OpenFile(path, os.O_RDONLY, 0644)
	defer file.Close()
	if err != nil {
		return make([]byte, 0), err
	}

	buf := make([]byte, 200)
	var i int
	i, err = file.Read(buf)
	if i == 0 {
		return []byte{}, err
	}
	return buf[:i], err
}
