// Confidential. Intellectual property of Pinely Holdings Pte. Ltd. Refer to CONFIDENTIAL file in the root for details

package neosync

import (
	"context"
	"time"
)

func Sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
