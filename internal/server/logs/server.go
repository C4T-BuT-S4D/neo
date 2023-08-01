package server

import (
	"context"
	"fmt"

	"github.com/c4t-but-s4d/neo/internal/server/common"
	"github.com/c4t-but-s4d/neo/internal/server/utils"
	"github.com/c4t-but-s4d/neo/pkg/gstream"
	logspb "github.com/c4t-but-s4d/neo/proto/go/logs"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/c4t-but-s4d/neo/internal/config"
)

const (
	logLinesBatchSize = 100
	maxMsgSize        = 4 * 1024 * 1024
)

func New(storage *LogStorage) *Server {
	return &Server{
		LoggingServer: common.NewLoggingServer("logs"),

		storage: storage,
	}
}

type Server struct {
	logspb.UnimplementedServiceServer
	common.LoggingServer

	storage *LogStorage
	config  *config.ExploitsConfig
}

func (s *Server) AddLogLines(ctx context.Context, lines *logspb.AddLogLinesRequest) (*emptypb.Empty, error) {
	s.GetMethodLogger(ctx).Infof("New request with %d lines", len(lines.Lines))

	decoded := make([]LogLine, 0, len(lines.Lines))
	for _, line := range lines.Lines {
		decoded = append(decoded, *NewLogLineFromProto(line))
	}
	if err := s.storage.Add(ctx, decoded); err != nil {
		return nil, utils.WrapErrorf(codes.Internal, "adding log lines: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) SearchLogLines(req *logspb.SearchLogLinesRequest, stream logspb.Service_SearchLogLinesServer) error {
	s.LogRequest(stream.Context(), req)

	opts := GetOptions{
		Exploit: req.Exploit,
		Version: req.Version,
	}
	lines, err := s.storage.Get(stream.Context(), opts)
	if err != nil {
		return utils.WrapErrorf(codes.Internal, "searching log lines: %v", err)
	}

	cache := gstream.NewDynamicSizeCache[*LogLine, logspb.SearchLogLinesResponse](
		stream,
		maxMsgSize,
		func(lines []*LogLine) (*logspb.SearchLogLinesResponse, error) {
			resp := &logspb.SearchLogLinesResponse{
				Lines: make([]*logspb.LogLine, 0, len(lines)),
			}
			for _, line := range lines {
				protoLine, err := line.ToProto()
				if err != nil {
					return nil, fmt.Errorf("converting line to proto: %w", err)
				}
				resp.Lines = append(resp.Lines, protoLine)
			}
			return resp, nil
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
