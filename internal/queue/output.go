package queue

import (
	"fmt"

	"neo/internal/models"
)

type Output struct {
	Exploit *models.Exploit
	Target  *models.Target
	Out     []byte
}

func NewOutput(job *Job, out []byte) *Output {
	return &Output{
		Exploit: job.Exploit,
		Target:  job.Target,
		Out:     out,
	}
}

func (o *Output) String() string {
	return fmt.Sprintf("%s (%s): %s", o.Exploit, o.Target, o.Out)
}
