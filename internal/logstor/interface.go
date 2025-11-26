package logstor

import (
	"context"
	"iter"

	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

type Storage interface {
	Add(ctx context.Context, lines ...*logspb.LogLine) error
	Search(ctx context.Context, req *logspb.SearchLogLinesRequest) iter.Seq2[*logspb.LogLine, error]
}
