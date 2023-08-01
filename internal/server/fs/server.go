package fs

import (
	"fmt"
	"os"

	"github.com/c4t-but-s4d/neo/internal/server/common"
	serverConfig "github.com/c4t-but-s4d/neo/internal/server/config"
	"github.com/c4t-but-s4d/neo/internal/server/utils"
	fspb "github.com/c4t-but-s4d/neo/proto/go/fileserver"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"

	"github.com/c4t-but-s4d/neo/internal/config"
	"github.com/c4t-but-s4d/neo/pkg/filestream"
)

func New(cfg *serverConfig.Config) (*Server, error) {
	fs, err := newOsFs(cfg.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("creating filesystem: %w", err)
	}
	ems := &Server{
		LoggingServer: common.NewLoggingServer("fileserver"),

		fs: fs,
	}
	return ems, nil
}

type Server struct {
	fspb.UnimplementedServiceServer
	common.LoggingServer

	config *config.ExploitsConfig
	fs     filesystem
}

func (s *Server) UploadFile(stream fspb.Service_UploadFileServer) error {
	info := &fspb.FileInfo{Uuid: uuid.NewString()}
	s.GetMethodLogger(stream.Context()).Infof("New file upload: %v", info)

	of, err := s.fs.Create(info.Uuid)
	if err != nil {
		return utils.WrapErrorf(codes.Internal, "Failed to create file: %v", err)
	}
	defer func() {
		if cerr := of.Close(); cerr != nil {
			err = utils.WrapErrorf(codes.Internal, "Failed to close output file")
		}
		if err != nil {
			if rerr := os.Remove(of.Name()); rerr != nil {
				logrus.Errorf("Error removing the file on error: %v", err)
			}
		}
	}()

	if err := filestream.Save(stream, of); err != nil {
		return utils.WrapErrorf(codes.Internal, "Failed to upload file from stream: %v", err)
	}
	if err := stream.SendAndClose(info); err != nil {
		return utils.WrapErrorf(codes.Internal, "Failed to send response & close connection: %v", err)
	}
	return nil
}

func (s *Server) DownloadFile(fi *fspb.FileInfo, stream fspb.Service_DownloadFileServer) error {
	s.LogRequest(stream.Context(), fi)

	f, err := s.fs.Open(fi.Uuid)
	if err != nil {
		return utils.WrapErrorf(codes.NotFound, "Failed to find file by uuid(%s): %v", fi.Uuid, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.Errorf("Error closing downloaded file: %v", err)
		}
	}()
	if err := filestream.Load(f, stream); err != nil {
		return utils.WrapErrorf(codes.NotFound, "Failed to find file by uuid(%s): %v", fi.Uuid, err)
	}
	return nil
}
