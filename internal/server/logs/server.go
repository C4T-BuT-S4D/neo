package logs

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/c4t-but-s4d/neo/v2/internal/logstor"
	"github.com/c4t-but-s4d/neo/v2/internal/server/logging"
	"github.com/c4t-but-s4d/neo/v2/pkg/gstream"
	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

const (
	maxMsgSize  = 4 * 1024 * 1024
	maxLogLines = 100
)

func New(storage logstor.Storage) *Server {
	return &Server{
		Server: logging.NewServer("logs"),

		storage: storage,
	}
}

type Server struct {
	logspb.UnimplementedServiceServer
	logging.Server

	storage logstor.Storage
}

func (s *Server) AddLogLines(ctx context.Context, request *logspb.AddLogLinesRequest) (*emptypb.Empty, error) {
	s.GetMethodLogger(ctx).Infof("New request with %d lines", len(request.GetLines()))

	if err := s.storage.Add(ctx, request.GetLines()...); err != nil {
		return nil, s.WrapErrorf(ctx, codes.Internal, "adding log lines: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) SearchLogLines(req *logspb.SearchLogLinesRequest, stream logspb.Service_SearchLogLinesServer) error {
	s.LogRequest(stream.Context(), req)

	linesIter := s.storage.Search(stream.Context(), req)

	cache := gstream.NewDynamicSizeCache(
		stream,
		maxMsgSize,
		maxLogLines,
		gstream.VTProtoSizer[*logspb.LogLine](),
		func(lines []*logspb.LogLine) (*logspb.SearchLogLinesResponse, error) {
			return &logspb.SearchLogLinesResponse{
				Lines: lines,
			}, nil
		},
	)

	for line, err := range linesIter {
		if err != nil {
			// Flush any remaining lines before returning error
			if flushErr := cache.Flush(); flushErr != nil {
				s.GetMethodLogger(stream.Context()).Errorf("flushing cache before error: %v", flushErr)
			}
			return s.WrapErrorf(stream.Context(), codes.Internal, "iterating log lines: %v", err)
		}

		if err := cache.Queue(line); err != nil {
			return s.WrapErrorf(stream.Context(), codes.Internal, "queueing log line: %v", err)
		}
	}

	if err := cache.Flush(); err != nil {
		return s.WrapErrorf(stream.Context(), codes.Internal, "flushing last batch: %v", err)
	}
	return nil
}
