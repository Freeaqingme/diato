package worker

import (
	"fmt"

	pb "diato/pb"
	"diato/util/stop"

	"google.golang.org/grpc"
)

func (w *Worker) rpcInit() error {
	conn, err := grpc.Dial(
		"127.0.0.1:2938",
		grpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("Could not connect to RPC server: %s", err.Error())
	}

	stop.NewStopper(func() {
		conn.Close()
	})

	w.userBackend = pb.NewUserBackendClient(conn)
	return nil
}
