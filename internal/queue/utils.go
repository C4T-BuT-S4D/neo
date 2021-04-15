package queue

import (
	"io"
)

const (
	chunkSize      = 32768
	smallChunkSize = 1024
)

type readCallback func([]byte)

// safeReadOutput reads from io.Reader, returning
// Output instances with possibly overlapping parts
// to prevent splitting of flags between reads.
func safeReadOutput(r io.Reader, cb readCallback) error {
	// reading into two buffers
	buffers := [][]byte{
		make([]byte, chunkSize),
		make([]byte, chunkSize),
	}

	// b is current buffer index
	for b := 0; ; b = (b + 1) % 2 {
		var (
			n, sn     int
			err, serr error
		)

		// second buffer index
		ot := (b + 1) % 2

		n, err = r.Read(buffers[b])
		// if the full buffer is read, it's possible for
		// the second part of the same data chunk to be next
		if n == chunkSize {
			sn, serr = r.Read(buffers[ot])
			if err == nil && serr != nil {
				err = serr
			}
		}
		if n > 0 {
			var res []byte
			var resLen int
			// both reads returned something, so we need to return
			// <current data> + <first small chunk of next data>
			if sn > 0 {
				resLen = n
				if sn > smallChunkSize {
					resLen += smallChunkSize
				} else {
					resLen += sn
				}
				res = make([]byte, resLen)
				copy(res, buffers[b])
				copy(res, buffers[ot][:resLen-n])
			} else {
				res = buffers[b]
				resLen = n
			}
			cb(res[:resLen])
		}
		if err != nil {
			return err
		}
	}
}
