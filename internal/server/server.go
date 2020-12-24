package server

import (
	"context"
	"io"
	"os"
	"path"
	"regexp"
	"sync"
	"time"

	"neo/internal/config"
	"neo/pkg/filestream"
	"neo/pkg/hostbucket"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	neopb "neo/lib/genproto/neo"
)

type fileInterface interface {
	io.ReadWriteCloser
	Name() string
}

type filesystem interface {
	Create(string) (fileInterface, error)
	Open(string) (fileInterface, error)
}

type osFs struct {
	baseDir string
}

func (o osFs) Create(f string) (fileInterface, error) {
	return os.Create(path.Join(o.baseDir, f))
}

func (o osFs) Open(f string) (fileInterface, error) {
	return os.Open(path.Join(o.baseDir, f))
}

func New(cfg *Config, storage *CachedStorage) *ExploitManagerServer {
	ems := &ExploitManagerServer{
		storage: storage,
		fs:      osFs{cfg.BaseDir},
		buckets: hostbucket.New(cfg.FarmConfig.Teams),
		visits:  newVisitsMap(),
	}
	ems.UpdateConfig(cfg)
	return ems
}

type ExploitManagerServer struct {
	neopb.UnimplementedExploitManagerServer
	storage  *CachedStorage
	config   *config.Config
	cfgMutex sync.RWMutex
	buckets  *hostbucket.HostBucket
	visits   *visitsMap
	fs       filesystem
}

func (em *ExploitManagerServer) UpdateConfig(cfg *Config) {
	em.cfgMutex.Lock()
	defer em.cfgMutex.Unlock()
	em.config = &config.Config{
		PingEvery:    cfg.PingEvery,
		RunEvery:     cfg.RunEvery,
		Timeout:      cfg.Timeout,
		FarmUrl:      cfg.FarmConfig.Url,
		FarmPassword: cfg.FarmConfig.Password,
		FlagRegexp:   regexp.MustCompile(cfg.FarmConfig.FlagRegexp),
	}
	em.buckets.UpdateTeams(cfg.FarmConfig.Teams)
}

func (em *ExploitManagerServer) UploadFile(stream neopb.ExploitManager_UploadFileServer) (err error) {
	info := &neopb.FileInfo{Uuid: uuid.New().String()}
	of, err := em.fs.Create(info.GetUuid())
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
	return stream.SendAndClose(info)
}

func (em *ExploitManagerServer) DownloadFile(fi *neopb.FileInfo, stream neopb.ExploitManager_DownloadFileServer) error {
	f, err := em.fs.Open(fi.GetUuid())
	if err != nil {
		return logErrorf(codes.NotFound, "Failed to find file by uuid(%s): %v", fi.GetUuid(), err)
	}
	defer f.Close()
	if err := filestream.Load(f, stream); err != nil {
		return logErrorf(codes.NotFound, "Failed to find file by uuid(%s): %v", fi.GetUuid(), err)
	}
	return nil
}

func (em *ExploitManagerServer) Exploit(_ context.Context, r *neopb.ExploitRequest) (*neopb.ExploitResponse, error) {
	state, ok := em.storage.State(r.GetExploitId())
	if !ok {
		return nil, logErrorf(codes.NotFound, "Failed to find an exploit state = %v", state.ExploitId)
	}
	cfg, ok := em.storage.Configuration(state)
	if !ok {
		return nil, logErrorf(codes.NotFound, "Failed to find an exploit configuration = %v", state.ExploitId)
	}
	return &neopb.ExploitResponse{
		State:  state,
		Config: cfg,
	}, nil
}

func (em *ExploitManagerServer) UpdateExploit(_ context.Context, r *neopb.UpdateExploitRequest) (*neopb.UpdateExploitResponse, error) {
	ns := &neopb.ExploitState{
		ExploitId: r.GetExploitId(),
		File:      r.GetFile(),
	}
	if err := em.storage.UpdateExploitVersion(ns, r.GetConfig()); err != nil {
		return nil, logErrorf(codes.Internal, "Failed to update exploit version: %v", err)
	}
	return &neopb.UpdateExploitResponse{
		State: ns,
	}, nil
}

func (em *ExploitManagerServer) Ping(_ context.Context, r *neopb.PingRequest) (*neopb.PingResponse, error) {
	logrus.Infof("Got %s from: %s", neopb.PingRequest_PingType_name[int32(r.GetType())], r.GetClientId())

	if r.Type == neopb.PingRequest_HEARTBEAT {
		em.cfgMutex.RLock()
		defer em.cfgMutex.RUnlock()
		em.visits.Add(r.GetClientId())
		em.buckets.AddNode(r.GetClientId(), int(r.GetWeight()))
	} else if r.Type == neopb.PingRequest_LEAVE {
		em.visits.MarkInvalid(r.GetClientId())
	}

	return &neopb.PingResponse{
		State: &neopb.ServerState{
			ClientTeamMap: em.buckets.Buckets(),
			Exploits:      em.storage.States(),
			Config:        config.ToProto(em.config),
		},
	}, nil
}

func (em *ExploitManagerServer) checkClients() {
	deadClients := em.visits.Invalidate(time.Now(), em.config.PingEvery)
	logrus.Infof("Heartbeat: got dead clients: %v", deadClients)
	for _, c := range deadClients {
		em.buckets.DeleteNode(c)
	}
}

func (em *ExploitManagerServer) HeartBeat(ctx context.Context) {
	ticker := time.NewTicker(em.config.PingEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			em.checkClients()
		case <-ctx.Done():
			return
		}
	}
}

func logErrorf(code codes.Code, fmt string, values ...interface{}) error {
	err := status.Errorf(code, fmt, values...)
	logrus.Errorf("%v", err)
	return err
}
