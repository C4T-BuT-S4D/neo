package server

import (
	"context"

	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/c4t-but-s4d/neo/v2/internal/logstor"
	"github.com/c4t-but-s4d/neo/v2/internal/server/common"
	"github.com/c4t-but-s4d/neo/v2/internal/server/utils"
	"github.com/c4t-but-s4d/neo/v2/pkg/gstream"
	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

const (
	logLinesBatchSize = 100
	maxMsgSize        = 4 * 1024 * 1024
)

func New(storage logstor.Storage) *Server {
	return &Server{
		LoggingServer: common.NewLoggingServer("logs"),

		storage: storage,
	}
}

type Server struct {
	logspb.UnimplementedServiceServer
	common.LoggingServer

	storage logstor.Storage
}

func (s *Server) AddLogLines(ctx context.Context, request *logspb.AddLogLinesRequest) (*emptypb.Empty, error) {
	s.GetMethodLogger(ctx).Infof("New request with %d lines", len(request.Lines))

	decoded := lo.Map(request.Lines, func(line *logspb.LogLine, _ int) *logstor.Line {
		return logstor.NewLineFromProto(line)
	})
	if err := s.storage.Add(ctx, decoded...); err != nil {
		return nil, utils.WrapErrorf(codes.Internal, "adding log lines: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) SearchLogLines(req *logspb.SearchLogLinesRequest, stream logspb.Service_SearchLogLinesServer) error {
	s.LogRequest(stream.Context(), req)

	var opts []logstor.SearchOption
	if req.Limit != 0 {
		opts = append(opts, logstor.SearchWithLimit(int(req.Limit)))
	}
	if req.LastToken != "" {
		opts = append(opts, logstor.SearchWithLastToken(req.LastToken))
	}

	lines, err := s.storage.Search(stream.Context(), req.Exploit, req.Version, opts...)
	if err != nil {
		return utils.WrapErrorf(codes.Internal, "searching log lines: %v", err)
	}

	cache := gstream.NewDynamicSizeCache[*logstor.Line, logspb.SearchLogLinesResponse](
		stream,
		maxMsgSize,
		func(lines []*logstor.Line) (*logspb.SearchLogLinesResponse, error) {
			return &logspb.SearchLogLinesResponse{
				Lines: lo.Map(lines, func(line *logstor.Line, _ int) *logspb.LogLine {
					return line.ToProto()
				}),
			}, nil
		},
	)

	for i, line := range lines {
		if err := cache.Queue(line); err != nil {
			return utils.WrapErrorf(codes.Internal, "queueing log line: %v", err)
		}
		if (i-1+logLinesBatchSize)%logLinesBatchSize == 0 {
			if err := cache.Flush(); err != nil {
				return utils.WrapErrorf(codes.Internal, "flushing batch: %v", err)
			}
		}
	}
	if err := cache.Flush(); err != nil {
		return utils.WrapErrorf(codes.Internal, "flushing last batch: %v", err)
	}
	return nil
}
