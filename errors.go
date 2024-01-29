package godimensionalflagfields

import "fmt"

type ErrOutOfBounds struct {
	boundType string
	bound     uint
	input     uint
}

func (e ErrOutOfBounds) Error() string {
	return fmt.Sprintf("Outside of bound %s. Max of %d but received %d", e.boundType, e.bound, e.input)
}
