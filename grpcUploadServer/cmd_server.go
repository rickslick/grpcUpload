package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"

	"github.com/urfave/cli"

	proto "github.com/rickslick/grpcUpload/proto"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server interface {
	Listen() (err error)
	Close()
}
type ServerGRPC struct {
	logger  zerolog.Logger
	server  *grpc.Server
	Address string

	certificate string
	key         string
	destDir     string
}

type ServerGRPCConfig struct {
	Certificate string
	Key         string
	Address     string
	DestDir     string
}

func NewServerGRPC(cfg ServerGRPCConfig) (s ServerGRPC, err error) {
	s.logger = zerolog.New(os.Stdout).
		With().
		Str("from", "server").
		Logger()

	if cfg.Address == "" {
		err = errors.Errorf("Address must be specified")
		s.logger.Error().Msg("Address must be specified")
		return
	}

	s.Address = cfg.Address
	s.certificate = cfg.Certificate
	s.key = cfg.Key

	if _, err = os.Stat(cfg.DestDir); err != nil {
		s.logger.Error().Msg("Directory doesnt exist")
		return
	}

	s.destDir = cfg.DestDir
	return
}

func (s *ServerGRPC) Listen() (err error) {
	var (
		listener  net.Listener
		grpcOpts  = []grpc.ServerOption{}
		grpcCreds credentials.TransportCredentials
	)

	listener, err = net.Listen("tcp", s.Address)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to listen on  %d",
			s.Address)
		return
	}

	if s.certificate != "" && s.key != "" {
		grpcCreds, err = credentials.NewServerTLSFromFile(
			s.certificate, s.key)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to create tls grpc server using cert %s and key %s",
				s.certificate, s.key)
			return
		}

		grpcOpts = append(grpcOpts, grpc.Creds(grpcCreds))
	}

	s.server = grpc.NewServer(grpcOpts...)
	proto.RegisterRkUploaderServiceServer(s.server, s)

	err = s.server.Serve(listener)
	if err != nil {
		err = errors.Wrapf(err, "errored listening for grpc connections")
		return
	}

	return
}

//writeToFp takes in a file pointer and byte array and writes the byte array into the file
//returns error if pointer is nil or error in writing to file
func writeToFp(fp *os.File, data []byte) error {
	w := 0
	n := len(data)
	for {

		nw, err := fp.Write(data[w:])
		if err != nil {
			return err
		}
		w += nw
		if nw >= n {
			return nil
		}
	}

}
func (s *ServerGRPC) UploadFile(stream proto.RkUploaderService_UploadFileServer) (err error) {
	firstChunk := true
	var fp *os.File

	var fileData *proto.UploadRequestType
	var filename string
	for {

		fileData, err = stream.Recv() //ignoring the data  TO-Do save files received

		if err != nil {
			if err == io.EOF {
				break
			}

			err = errors.Wrapf(err,
				"failed unexpectadely while reading chunks from stream")
			return
		}

		if firstChunk { //first chunk contains file name

			if fileData.Filename != "" { //create file

				fp, err = os.Create(path.Join(s.destDir, filepath.Base(fileData.Filename)))

				if err != nil {
					s.logger.Error().Msg("Unable to create file  :" + fileData.Filename)
					stream.SendAndClose(&proto.UploadResponseType{
						Message: "Unable to create file :" + fileData.Filename,
						Code:    proto.UploadStatusCode_Failed,
					})
					return
				}
				defer fp.Close()
			} else {
				s.logger.Error().Msg("FileName not provided in first chunk  :" + fileData.Filename)
				stream.SendAndClose(&proto.UploadResponseType{
					Message: "FileName not provided in first chunk:" + fileData.Filename,
					Code:    proto.UploadStatusCode_Failed,
				})
				return

			}
			filename = fileData.Filename
			firstChunk = false
		}

		err = writeToFp(fp, fileData.Content)
		if err != nil {
			s.logger.Error().Msg("Unable to write chunk of filename :" + fileData.Filename + " " + err.Error())
			stream.SendAndClose(&proto.UploadResponseType{
				Message: "Unable to write chunk of filename :" + fileData.Filename,
				Code:    proto.UploadStatusCode_Failed,
			})
			return
		}
	}

	//s.logger.Info().Msg("upload received")
	err = stream.SendAndClose(&proto.UploadResponseType{
		Message: "Upload received with success",
		Code:    proto.UploadStatusCode_Ok,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"failed to send status code")
		return
	}
	fmt.Println("Successfully received and stored the file :" + filename + " in " + s.destDir)
	return
}

func (s *ServerGRPC) Close() {
	if s.server != nil {
		s.server.Stop()
	}

	return
}

func StartServerCommand() cli.Command {

	return cli.Command{
		Name:  "serve",
		Usage: "initiates a gRPC server",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "a",
				Usage: "Address to listen",
				Value: "localhost:80",
			},

			&cli.StringFlag{
				Name:  "key",
				Usage: "path to TLS certificate",
			},
			&cli.StringFlag{
				Name:  "certificate",
				Usage: "path to TLS certificate",
			},
			&cli.StringFlag{
				Name:  "d",
				Usage: "Destrination directory Default is /tmp",
				Value: "/tmp",
			},
		},
		Action: func(c *cli.Context) error {
			grpcServer, err := NewServerGRPC(ServerGRPCConfig{
				Address:     c.String("a"),
				Certificate: c.String("certificate"),
				Key:         c.String("key"),
				DestDir:     c.String("d"),
			})
			if err != nil {
				fmt.Println("error is creating server")

				return err
			}
			server := &grpcServer
			err = server.Listen()

			defer server.Close()
			return nil
		},
	}

}
