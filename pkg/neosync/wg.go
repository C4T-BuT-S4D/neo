// Confidential. Intellectual property of Pinely Holdings Pte. Ltd. Refer to CONFIDENTIAL file in the root for details

package neosync

import (
	"sync"
)

type WaitGroup struct {
	sync.WaitGroup
}

func NewWG() WaitGroup {
	return WaitGroup{}
}

func (wg *WaitGroup) Await() <-chan struct{} {
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()
	return c
}
