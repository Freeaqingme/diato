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

package modsec

import (
	pb "diato/module/modsec/pb"

	empty "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func (m *module) RegisterRpcEndpoints(s *grpc.Server) {
	pb.RegisterModuleModsecServer(s, &rpcServer{m})
}

type rpcServer struct {
	module *module
}

func (s *rpcServer) GetRules(ctx context.Context, _ *empty.Empty) (*pb.RuleSets, error) {
	return s.module.rules, nil
}
