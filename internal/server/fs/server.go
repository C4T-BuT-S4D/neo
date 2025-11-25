package fs

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"

	serverConfig "github.com/c4t-but-s4d/neo/v2/internal/server/config"
	"github.com/c4t-but-s4d/neo/v2/internal/server/logging"
	"github.com/c4t-but-s4d/neo/v2/pkg/filestream"
	fspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/fileserver"
)

func New(cfg *serverConfig.Config) (*Server, error) {
	fs, err := newOsFs(cfg.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("creating filesystem: %w", err)
	}
	ems := &Server{
		Server: logging.NewServer("fileserver"),

		fs: fs,
	}
	return ems, nil
}

type Server struct {
	fspb.UnimplementedServiceServer
	logging.Server

	fs filesystem
}

func (s *Server) UploadFile(stream fspb.Service_UploadFileServer) error {
	info := &fspb.FileInfo{Uuid: uuid.NewString()}
	s.GetMethodLogger(stream.Context()).Infof("New file upload: %v", info)

	of, err := s.fs.Create(info.GetUuid())
	if err != nil {
		return s.WrapErrorf(stream.Context(), codes.Internal, "creating file: %v", err)
	}
	defer func() {
		if cerr := of.Close(); cerr != nil {
			err = s.WrapErrorf(stream.Context(), codes.Internal, "closing output file: %v", cerr)

			if rerr := os.Remove(of.Name()); rerr != nil {
				s.GetMethodLogger(stream.Context()).Errorf("removing the file on error: %v", rerr)
			}
		}
	}()

	if err := filestream.Save(stream, of); err != nil {
		return s.WrapErrorf(stream.Context(), codes.Internal, "uploading file from stream: %v", err)
	}
	if err := stream.SendAndClose(info); err != nil {
		return s.WrapErrorf(stream.Context(), codes.Internal, "sending response & closing connection: %v", err)
	}
	return nil
}

func (s *Server) DownloadFile(fi *fspb.FileInfo, stream fspb.Service_DownloadFileServer) error {
	s.LogRequest(stream.Context(), fi)

	f, err := s.fs.Open(fi.GetUuid())
	if err != nil {
		return s.WrapErrorf(stream.Context(), codes.NotFound, "finding file by uuid(%s): %v", fi.GetUuid(), err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			s.GetMethodLogger(stream.Context()).Errorf("closing downloaded file: %v", err)
		}
	}()
	if err := filestream.Load(f, stream); err != nil {
		return s.WrapErrorf(stream.Context(), codes.NotFound, "loading file: %v", err)
	}
	return nil
}
