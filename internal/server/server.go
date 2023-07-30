package server

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"neo/pkg/gstream"
	"neo/pkg/pubsub"

	"neo/internal/config"
	"neo/pkg/filestream"
	"neo/pkg/hostbucket"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	neopb "neo/lib/genproto/neo"
)

const (
	logLinesBatchSize = 100
	maxMsgSize        = 4 * 1024 * 1024
)

func New(cfg *Config, storage *CachedStorage, logStore *LogStorage) (*ExploitManagerServer, error) {
	fs, err := newOsFs(cfg.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("creating filesystem: %w", err)
	}
	ems := &ExploitManagerServer{
		storage:         storage,
		fs:              fs,
		buckets:         hostbucket.New(cfg.FarmConfig.Teams),
		visits:          newVisitsMap(),
		singleRunPubSub: pubsub.NewPubSub[*neopb.SingleRunRequest](),
		broadcastPubSub: pubsub.NewPubSub[*neopb.Command](),
		logStorage:      logStore,
		metrics:         NewMetrics(),
	}
	ems.UpdateConfig(cfg)
	return ems, nil
}

type ExploitManagerServer struct {
	neopb.UnimplementedExploitManagerServer
	storage    *CachedStorage
	logStorage *LogStorage
	config     *config.Config
	cfgMutex   sync.RWMutex
	buckets    *hostbucket.HostBucket
	visits     *visitsMap
	fs         filesystem
	metrics    *Metrics

	singleRunPubSub *pubsub.PubSub[*neopb.SingleRunRequest]
	broadcastPubSub *pubsub.PubSub[*neopb.Command]
}

func (s *ExploitManagerServer) UpdateConfig(cfg *Config) {
	env := make([]string, 0, len(cfg.Environ))
	for k, v := range cfg.Environ {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	s.cfgMutex.Lock()
	defer s.cfgMutex.Unlock()
	s.config = &config.Config{
		PingEvery:    cfg.PingEvery,
		SubmitEvery:  cfg.SubmitEvery,
		FarmURL:      cfg.FarmConfig.URL,
		FarmPassword: cfg.FarmConfig.Password,
		FlagRegexp:   regexp.MustCompile(cfg.FarmConfig.FlagRegexp),
		Environ:      env,
	}
	s.buckets.UpdateTeams(cfg.FarmConfig.Teams)
}

func (s *ExploitManagerServer) Ping(_ context.Context, r *neopb.PingRequest) (*neopb.PingResponse, error) {
	logrus.Infof("Got %s from: %s", neopb.PingRequest_PingType_name[int32(r.Type)], r.ClientId)

	if r.Type == neopb.PingRequest_HEARTBEAT {
		s.cfgMutex.RLock()
		defer s.cfgMutex.RUnlock()
		s.visits.Add(r.ClientId)
		s.buckets.AddNode(r.ClientId, int(r.Weight))
	} else if r.Type == neopb.PingRequest_LEAVE {
		s.visits.MarkInvalid(r.ClientId)
	}

	return &neopb.PingResponse{
		State: &neopb.ServerState{
			ClientTeamMap: s.buckets.Buckets(),
			Exploits:      s.storage.States(),
			Config:        config.ToProto(s.config),
		},
	}, nil
}

func (s *ExploitManagerServer) UploadFile(stream neopb.ExploitManager_UploadFileServer) error {
	info := &neopb.FileInfo{Uuid: uuid.NewString()}
	of, err := s.fs.Create(info.Uuid)
	if err != nil {
		return logErrorf(codes.Internal, "Failed to create file: %v", err)
	}
	defer func() {
		if cerr := of.Close(); cerr != nil {
			err = logErrorf(codes.Internal, "Failed to close output file")
		}
		if err != nil {
			if rerr := os.Remove(of.Name()); rerr != nil {
				logrus.Errorf("Error removing the file on error: %v", err)
			}
		}
	}()

	if err := filestream.Save(stream, of); err != nil {
		return logErrorf(codes.Internal, "Failed to upload file from stream: %v", err)
	}
	if err := stream.SendAndClose(info); err != nil {
		return logErrorf(codes.Internal, "Failed to send response & close connection: %v", err)
	}
	return nil
}

func (s *ExploitManagerServer) DownloadFile(fi *neopb.FileInfo, stream neopb.ExploitManager_DownloadFileServer) error {
	f, err := s.fs.Open(fi.Uuid)
	if err != nil {
		return logErrorf(codes.NotFound, "Failed to find file by uuid(%s): %v", fi.Uuid, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.Errorf("Error closing downloaded file: %v", err)
		}
	}()
	if err := filestream.Load(f, stream); err != nil {
		return logErrorf(codes.NotFound, "Failed to find file by uuid(%s): %v", fi.Uuid, err)
	}
	return nil
}

func (s *ExploitManagerServer) Exploit(_ context.Context, r *neopb.ExploitRequest) (*neopb.ExploitResponse, error) {
	state, ok := s.storage.GetState(r.ExploitId)
	if !ok {
		return nil, logErrorf(codes.NotFound, "Failed to find an exploit state = %v", state.ExploitId)
	}
	return &neopb.ExploitResponse{
		State: state,
	}, nil
}

func (s *ExploitManagerServer) UpdateExploit(_ context.Context, r *neopb.UpdateExploitRequest) (*neopb.UpdateExploitResponse, error) {
	newState, err := s.storage.UpdateExploitVersion(r.State)
	if err != nil {
		return nil, logErrorf(codes.Internal, "Failed to update exploit version: %v", err)
	}
	return &neopb.UpdateExploitResponse{State: newState}, nil
}

func (s *ExploitManagerServer) BroadcastCommand(_ context.Context, r *neopb.Command) (*emptypb.Empty, error) {
	logrus.Infof("Received broadcast request to run %v", r)
	s.broadcastPubSub.Publish(r)
	return &emptypb.Empty{}, nil
}

func (s *ExploitManagerServer) BroadcastRequests(_ *emptypb.Empty, stream neopb.ExploitManager_BroadcastRequestsServer) error {
	sub := s.broadcastPubSub.Subscribe(stream.Send)
	defer s.broadcastPubSub.Unsubscribe(sub)

	sub.Run(stream.Context())
	return nil
}

func (s *ExploitManagerServer) SingleRun(_ context.Context, r *neopb.SingleRunRequest) (*emptypb.Empty, error) {
	logrus.Infof("Received single run request %v", r)
	s.singleRunPubSub.Publish(r)
	return &emptypb.Empty{}, nil
}

func (s *ExploitManagerServer) SingleRunRequests(_ *emptypb.Empty, stream neopb.ExploitManager_SingleRunRequestsServer) error {
	sub := s.singleRunPubSub.Subscribe(stream.Send)
	defer s.singleRunPubSub.Unsubscribe(sub)

	sub.Run(stream.Context())
	return nil
}

func (s *ExploitManagerServer) AddLogLines(ctx context.Context, lines *neopb.AddLogLinesRequest) (*emptypb.Empty, error) {
	decoded := make([]LogLine, 0, len(lines.Lines))
	for _, line := range lines.Lines {
		decoded = append(decoded, *NewLogLineFromProto(line))
	}
	if err := s.logStorage.Add(ctx, decoded); err != nil {
		return nil, logErrorf(codes.Internal, "adding log lines: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *ExploitManagerServer) SearchLogLines(req *neopb.SearchLogLinesRequest, stream neopb.ExploitManager_SearchLogLinesServer) error {
	opts := GetOptions{
		Exploit: req.Exploit,
		Version: req.Version,
	}
	lines, err := s.logStorage.Get(stream.Context(), opts)
	if err != nil {
		return logErrorf(codes.Internal, "searching log lines: %v", err)
	}

	cache := gstream.NewDynamicSizeCache[*LogLine, neopb.SearchLogLinesResponse](
		stream,
		maxMsgSize,
		func(lines []*LogLine) (*neopb.SearchLogLinesResponse, error) {
			resp := &neopb.SearchLogLinesResponse{
				Lines: make([]*neopb.LogLine, 0, len(lines)),
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
			return logErrorf(codes.Internal, "queueing log line: %v", err)
		}
		if (i-1+logLinesBatchSize)%logLinesBatchSize == 0 {
			if err := cache.Flush(); err != nil {
				return logErrorf(codes.Internal, "flushing batch: %v", err)
			}
		}
	}
	if err := cache.Flush(); err != nil {
		return logErrorf(codes.Internal, "flushing last batch: %v", err)
	}
	return nil
}

func (s *ExploitManagerServer) HeartBeat(ctx context.Context) {
	ticker := time.NewTicker(s.config.PingEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.checkClients()
		case <-ctx.Done():
			return
		}
	}
}

func (s *ExploitManagerServer) UpdateMetrics(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.updateMetrics()
		case <-ctx.Done():
			return
		}
	}
}

func (s *ExploitManagerServer) checkClients() {
	alive, dead := s.visits.Invalidate(time.Now(), s.config.PingEvery)
	logrus.Infof("Heartbeat: got dead clients: %v, alive clients: %v", dead, alive)
	for _, c := range dead {
		s.buckets.DeleteNode(c)
	}
}

func (s *ExploitManagerServer) updateMetrics() {
	s.metrics.AliveClients.Set(float64(s.visits.Size()))
}

func logErrorf(code codes.Code, fmt string, values ...interface{}) error {
	err := status.Errorf(code, fmt, values...)
	logrus.Errorf("%v", err)
	return err // nolint
}
