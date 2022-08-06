package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/gjkim42/s3-fuse/pkg/s3"
	"github.com/hanwen/go-fuse/v2/fs"
)

var (
	endpoint = os.Getenv("ENDPOINT")
	region   = os.Getenv("REGION")
	bucket   = os.Getenv("BUCKET")
	debug    bool
)

func init() {
	flag.StringVar(&endpoint, "endpoint", endpoint, "S3 endpoint")
	flag.StringVar(&region, "region", region, "S3 region")
	flag.StringVar(&bucket, "bucket", bucket, "S3 bucket")
	flag.BoolVar(&debug, "debug", debug, "print debug data")
}

func setupSignalHandler() <-chan os.Signal {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	return stop
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  hello MOUNTPOINT")
	}

	stop := setupSignalHandler()

	opts := &fs.Options{}
	opts.Debug = debug

	server, err := fs.Mount(flag.Arg(0), s3.NewS3INode(endpoint, region, bucket), opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	defer func() {
		err := server.Unmount()
		if err != nil {
			log.Fatalf("Unmount fail: %v\n", err)
		}
	}()

	<-stop
}
