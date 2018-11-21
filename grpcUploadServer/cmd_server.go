package main

import (
	"fmt"
	"io"
	"net"
	"os"

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
}

type ServerGRPCConfig struct {
	Certificate string
	Key         string
	Address     string
}

func NewServerGRPC(cfg ServerGRPCConfig) (s ServerGRPC, err error) {
	s.logger = zerolog.New(os.Stdout).
		With().
		Str("from", "server").
		Logger()

	if cfg.Address == "" {
		err = errors.Errorf("Address must be specified")
		return
	}

	s.Address = cfg.Address
	s.certificate = cfg.Certificate
	s.key = cfg.Key

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

func (s *ServerGRPC) UploadFile(stream proto.RkUploaderService_UploadFileServer) (err error) {
	for {
		_, err = stream.Recv() //ignoring the data  TO-Do save files received
		if err != nil {
			if err == io.EOF {
				goto END
			}

			err = errors.Wrapf(err,
				"failed unexpectadely while reading chunks from stream")
			return
		}
	}

END:
	s.logger.Info().Msg("upload received")
	fmt.Println("upload received")
	err = stream.SendAndClose(&proto.UploadResponseType{
		Message: "Upload received with success",
		Code:    proto.UploadStatusCode_Ok,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"failed to send status code")
		return
	}

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
		},
		Action: func(c *cli.Context) error {
			grpcServer, err := NewServerGRPC(ServerGRPCConfig{
				Address:     c.String("a"),
				Certificate: c.String("certificate"),
				Key:         c.String("key"),
			})
			if err != nil {
				return err
			}
			server := &grpcServer
			err = server.Listen()

			defer server.Close()
			return nil
		},
	}

}
