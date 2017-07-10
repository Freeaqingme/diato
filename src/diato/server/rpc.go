// Diato - Reverse Proxying for Hipsters
//
// Copyright 2016-2017 Dolf Schimmel
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package server

import (
	"log"
	"net"

	pb "diato/pb"
	"diato/util/stop"

	empty "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func (s *Server) startRpc() error {
	// Ideally this should be a socket too, but there appears
	// to be no way to convert an FD into a net.conn object,
	// like there is for listeners.
	ln, err := net.Listen("tcp", "127.0.0.1:2938")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserBackendServer(grpcServer, &rpcUserBackendServer{s})
	pb.RegisterServerServer(grpcServer, &rpcServerServer{s})
	for _, module := range s.modules.modules {
		module.RegisterRpcEndpoints(grpcServer)
	}

	reflection.Register(grpcServer)

	stopper := stop.NewStopper(func() {
		ln.Close()
	})

	go func() {
		err := grpcServer.Serve(ln)
		if !stopper.IsStopping() {
			log.Fatalf("failed to serve: %v", err)
		}

		log.Print("RPC server stopped (sort of?) gracefully")
	}()

	return nil
}

type rpcServerServer struct {
	diato *Server
}

func (s *rpcServerServer) GetConfigContents(ctx context.Context, _ *empty.Empty) (*pb.ConfigContents, error) {
	return &pb.ConfigContents{s.diato.configFileContents}, nil
}

type rpcUserBackendServer struct {
	diato *Server
}

func (s *rpcUserBackendServer) GetServerForUser(ctx context.Context, in *pb.UserBackendRequest) (*pb.UserBackendResponse, error) {
	host, port, err := s.diato.userBackend.GetServerForUser(in.Name)
	return &pb.UserBackendResponse{host, port}, err
}
