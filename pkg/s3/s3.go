package s3

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type s3INode struct {
	fs.Inode
	svc    *s3.S3
	bucket string
}

var _ = (fs.NodeOnAdder)(&s3INode{})
var _ = (fs.NodeGetattrer)(&s3INode{})

func NewS3INode(endpoint, region, bucket string) *s3INode {
	trans := &http.Transport{
		TLSClientConfig: &tls.Config{
			// Workaround for the oracle cloud
			InsecureSkipVerify: true,
		},
	}

	config := aws.NewConfig().
		WithRegion(region).
		WithEndpoint(endpoint).
		WithCredentials(credentials.NewEnvCredentials()).
		WithHTTPClient(&http.Client{
			Transport: trans,
		}).
		WithS3ForcePathStyle(true)

	// All clients require a Session. The Session provides the client with
	// shared configuration such as region, endpoint, and credentials. A
	// Session should be shared where possible to take advantage of
	// configuration and credential caching. See the session package for
	// more information.
	sess := session.Must(session.NewSession(config))

	// Create a new instance of the service's client with a Session.
	// Optional aws.Config values can also be provided as variadic arguments
	// to the New function. This option allows you to provide service
	// specific configuration.
	svc := s3.New(sess)

	return &s3INode{
		svc:    svc,
		bucket: bucket,
	}
}

func (s *s3INode) OnAdd(ctx context.Context) {
	out, err := s.svc.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ListObjectsV2 error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "successfully listed objects in bucket %q\n", s.bucket)

	for _, obj := range out.Contents {
		out, err := s.svc.GetObjectWithContext(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(*obj.Key),
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
				// If the SDK can determine the request or retry delay was canceled
				// by a context the CanceledErrorCode error code will be returned.
				fmt.Fprintf(os.Stderr, "download canceled due to timeout, %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "failed to download object, %v\n", err)
			}
			os.Exit(1)
		}
		defer out.Body.Close()

		body, err := io.ReadAll(out.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read object, %v\n", err)
			os.Exit(1)
		}

		node := s.NewPersistentInode(
			ctx, &fs.MemRegularFile{
				Data: body,
				Attr: fuse.Attr{
					Mode: 0644,
				},
			}, fs.StableAttr{Ino: 2})
		s.AddChild(*obj.Key, node, false)
		fmt.Fprintf(os.Stderr, "successfully downloaded file %q/%q\n", s.bucket, *obj.Key)
	}
}

func (s *s3INode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755
	return 0
}
