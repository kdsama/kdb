/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package consensus

import (
	"context"
	"log"

	pb "github.com/kdsama/kdb/consensus/protodata"
)

// server is used to implement helloworld.GreeterServer.
type Server struct {
	pb.UnimplementedConsensusServer
}

// Acknowledgement that the heartbeat has been received
func (s *Server) Ack(ctx context.Context, in *pb.Hearbeat) (*pb.HearbeatResponse, error) {

	log.Printf("Received: %v", in.GetMessage())
	return &pb.HearbeatResponse{Message: "Hello " + in.GetMessage()}, nil
}

// SayHello implements helloworld.GreeterServer
func (s *Server) SendRecord(ctx context.Context, in *pb.WalEntry) (*pb.WalResponse, error) {

	log.Printf("Received: %v", in.GetEntry())
	return &pb.WalResponse{Message: "ok"}, nil
}
