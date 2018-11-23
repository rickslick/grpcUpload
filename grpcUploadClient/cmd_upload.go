package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	proto "github.com/rickslick/grpcUpload/proto"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/cheggaaa/pb.v1"
)

const chunkSize = 64 * 1024 // 64 KiB

type uploader struct {
	dir      string
	client   proto.RkUploaderServiceClient
	ctx      context.Context
	wg       sync.WaitGroup
	requests chan string // each request is a filepath on client accessible to client
	pool     *pb.Pool
}

func NewUploader(ctx context.Context, client proto.RkUploaderServiceClient, dir string) *uploader {
	d := &uploader{
		ctx:      ctx,
		client:   client,
		dir:      dir,
		requests: make(chan string),
	}
	for i := 0; i < 5; i++ {
		d.wg.Add(1)
		go d.worker()
	}
	d.pool, _ = pb.StartPool()
	return d
}

func (d *uploader) Stop() {
	close(d.requests)
	d.wg.Wait()
	d.pool.RefreshRate = 500 * time.Millisecond
	d.pool.Stop()
}

func (d *uploader) worker() {
	defer d.wg.Done()
	var (
		buf        []byte
		firstChunk bool
	)
	for request := range d.requests {

		//open file
		file, errOpen := os.Open(request)
		if errOpen != nil {
			errOpen = errors.Wrapf(errOpen,
				"failed to open file %s",
				request)
			return
		}

		defer file.Close()

		//start uploader
		streamUploader, err := d.client.UploadFile(d.ctx)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to create upload stream for file %s",
				request)
			return
		}
		defer streamUploader.CloseSend()
		stat, errstat := file.Stat()
		if errstat != nil {
			err = errors.Wrapf(err,
				"Unable to get file size  %s",
				request)
			return
		}
		//start progress bar
		bar := pb.New64(stat.Size()).Postfix(" " + filepath.Base(request))
		bar.Units = pb.U_BYTES
		d.pool.Add(bar)

		//create a buffer of chunkSize to be streamed
		buf = make([]byte, chunkSize)
		firstChunk = true
		for {
			n, errRead := file.Read(buf)
			if errRead != nil {
				if errRead == io.EOF {
					errRead = nil
					break
				}

				errRead = errors.Wrapf(errRead,
					"errored while copying from file to buf")
				return
			}
			if firstChunk {
				err = streamUploader.Send(&proto.UploadRequestType{
					Content:  buf[:n],
					Filename: request,
				})
				firstChunk = false
			} else {
				err = streamUploader.Send(&proto.UploadRequestType{
					Content: buf[:n],
				})
			}
			if err != nil {

				bar.FinishPrint("failed to send chunk via stream file : " + request)
				break
				//bar.Reset(0)
				//return
			}

			bar.Add64(int64(n))
		}
		status, err := streamUploader.CloseAndRecv()
		if err != nil { //retry needed

			fmt.Println("failed to receive upstream status response")
			return
		}

		if status.Code != proto.UploadStatusCode_Ok { //retry needed

			bar.FinishPrint("Error uploading file : " + request + " :" + status.Message)
			bar.Reset(0)
			return
		}

		bar.Finish()

	}

}

func (d *uploader) Do(filepath string) {
	d.requests <- filepath
}

func UploadFiles(ctx context.Context, client proto.RkUploaderServiceClient, filepathlist []string, dir string) error {

	d := NewUploader(ctx, client, dir)

	if dir != "" {

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if !file.IsDir() {

				d.Do(dir + "/" + file.Name())

			}
		}
	} else {

		for _, file := range filepathlist {
			d.Do(file)
		}
	}

	d.Stop()

	return nil
}

func uploadCommand() cli.Command {
	return cli.Command{
		Name:  "upload",
		Usage: "Uplooads files from server in parallel",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "a",
				Value: "localhost:port",
				Usage: "server address",
			},
			cli.StringFlag{
				Name:  "d",
				Value: ".",
				Usage: "base directory",
			},
			cli.StringFlag{
				Name:  "tls-path",
				Value: "",
				Usage: "directory to the TLS server.crt file",
			},
		},
		Action: func(c *cli.Context) error {
			options := []grpc.DialOption{}
			if p := c.String("tls-path"); p != "" {
				creds, err := credentials.NewClientTLSFromFile(
					filepath.Join(p, "server.crt"),
					"")
				if err != nil {
					log.Println(err)
					return err
				}
				options = append(options, grpc.WithTransportCredentials(creds))
			} else {
				options = append(options, grpc.WithInsecure())
			}
			addr := c.String("a")

			conn, err := grpc.Dial(addr, options...)
			if err != nil {
				log.Fatalf("cannot connect: %v", err)
			}
			defer conn.Close()

			return UploadFiles(context.Background(), proto.NewRkUploaderServiceClient(conn), []string{}, c.String("d"))
		},
	}
}
