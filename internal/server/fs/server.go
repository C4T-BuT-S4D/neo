package fs

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
		return logging.WrapErrorf(codes.Internal, "Failed to create file: %v", err)
	}
	defer func() {
		if cerr := of.Close(); cerr != nil {
			err = logging.WrapErrorf(codes.Internal, "Failed to close output file")
		}
		if err != nil {
			if rerr := os.Remove(of.Name()); rerr != nil {
				logrus.Errorf("Error removing the file on error: %v", err)
			}
		}
	}()

	if err := filestream.Save(stream, of); err != nil {
		return logging.WrapErrorf(codes.Internal, "Failed to upload file from stream: %v", err)
	}
	if err := stream.SendAndClose(info); err != nil {
		return logging.WrapErrorf(codes.Internal, "Failed to send response & close connection: %v", err)
	}
	return nil
}

func (s *Server) DownloadFile(fi *fspb.FileInfo, stream fspb.Service_DownloadFileServer) error {
	s.LogRequest(stream.Context(), fi)

	f, err := s.fs.Open(fi.GetUuid())
	if err != nil {
		return logging.WrapErrorf(codes.NotFound, "Failed to find file by uuid(%s): %v", fi.GetUuid(), err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.Errorf("Error closing downloaded file: %v", err)
		}
	}()
	if err := filestream.Load(f, stream); err != nil {
		return logging.WrapErrorf(codes.NotFound, "Failed to find file by uuid(%s): %v", fi.GetUuid(), err)
	}
	return nil
}
