package main

import (
	"context"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	v1 "github.com/stackrox/rox/generated/api/v1"
	"github.com/stackrox/rox/roxctl/common"
	"github.com/stackrox/rox/roxctl/common/auth"
	roxctlIO "github.com/stackrox/rox/roxctl/common/io"
	"github.com/stackrox/rox/roxctl/common/logger"
	"github.com/stackrox/rox/roxctl/common/printer"
	"io"
	"time"
)

func main() {
	defaultIO := roxctlIO.DefaultIO()
	conn, err := common.GetGRPCConnection(auth.TokenAuth(), logger.NewLogger(defaultIO, printer.DefaultColorPrinter()))
	if err != nil {
		panic(errors.Wrap(err, "could not establish gRPC connection to central"))
	}
	svc := v1.NewDeploymentServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	client, err := svc.ExportDeployments(ctx, &v1.ExportDeploymentRequest{})
	if err != nil {
		panic(errors.Wrap(err, "could not initialize stream client"))
	}

	marshaler := &jsonpb.Marshaler{}
	for {
		deployment, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			panic(errors.Wrap(err, "stream broken by unexpected error"))
		}
		if err := marshaler.Marshal(defaultIO.Out(), deployment); err != nil {
			panic(errors.Wrap(err, "unable to serialize deployment"))
		}
	}
}
