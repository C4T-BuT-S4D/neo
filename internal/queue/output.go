package queue

import (
	"fmt"
)

type Output struct {
	Name string
	Team string
	Out  []byte
}

func (o *Output) String() string {
	return fmt.Sprintf("%s (%s): %s", o.Name, o.Name, o.Out)
}
