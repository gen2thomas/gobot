package adaptors

import (
	"fmt"
	"slices"
)

type BusNumberValidator struct {
	validNumbers []int
}

// NewBusNumberValidator creates a new instance for a bus number validator, used for I2C and SPI.
func NewBusNumberValidator(validNumbers []int) *BusNumberValidator {
	return &BusNumberValidator{validNumbers: validNumbers}
}

func (bnv *BusNumberValidator) Validate(busNr int) error {
	if slices.Contains(bnv.validNumbers, busNr) {
		return nil
	}

	return fmt.Errorf("bus number %d out of range %v", busNr, bnv.validNumbers)
}
