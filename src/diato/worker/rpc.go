package worker

import (
	"context"
	"fmt"

	pb "diato/pb"
	"diato/util/stop"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

func (w *Worker) rpcInit() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(
		"127.0.0.1:2938",
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to RPC server: %s", err.Error())
	}

	stop.NewStopper(func() {
		conn.Close()
	})

	w.userBackend = pb.NewUserBackendClient(conn)
	return conn, nil
}

func (w *Worker) GetGrpcClientConn() *grpc.ClientConn {
	return w.grpcClientConn
}

func (w *Worker) getConfigContents() ([]byte, error) {
	// If we use the ServerClient on more than one
	// place make sure to store it somewhere centrally
	conf, err := pb.NewServerClient(w.grpcClientConn).GetConfigContents(context.Background(), &empty.Empty{})
	if err != nil {
		return []byte{}, fmt.Errorf("Could not retrieve config contents: %s", err.Error())
	}

	return conf.Contents, nil
}
