package neoproc

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// SetSubreaper sets current process as the subreaper for its children,
// i.e. children a re-parented to current process, not the init (pid=1).
func SetSubreaper() error {
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("calling prctl: %w", err)
	}
	return nil
}
