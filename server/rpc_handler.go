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
package server

import (
	"context"

	"github.com/kdsama/kdb/config"
	pb "github.com/kdsama/kdb/protodata"
)

type Handler struct {
	pb.UnimplementedConsensusServer
	server *Server
}

func NewHandler(server *Server) *Handler {
	return &Handler{}
}

// Acknowledgement that the heartbeat has been received
func (s *Handler) Ack(ctx context.Context, in *pb.Hearbeat) (*pb.HearbeatResponse, error) {

	// log.Printf("Received: %v", in.GetMessage())
	return &pb.HearbeatResponse{Message: "Hello " + in.GetMessage()}, nil
}

// Record received, now commit/ acknowledge according to the type of data
func (s *Handler) SendRecord(ctx context.Context, in *pb.WalEntry) (*pb.WalResponse, error) {

	switch in.Status {
	case int32(config.Acknowledge):
		// a function is required to just add a wal entry

		s.server.AcknowledgeRecord(&in.Entry)

	case int32(config.Commit):

		s.server.SetRecord(&in.Entry)
	}

	return &pb.WalResponse{Message: "ok"}, nil
}
func (s *Handler) Get(ctx context.Context, in *pb.GetKey) (*pb.GetResponse, error) {

	val, err := s.server.Get(in.Key)
	if err != nil {
		return &pb.GetResponse{Value: ""}, err
	}
	return &pb.GetResponse{Value: val}, nil
}

func (s *Handler) Broadcast(ctx context.Context, in *pb.BroadcastNode) (*pb.BroadcastNodeResponse, error) {

	// ADD NODE TO EXISTING NODES

	return &pb.BroadcastNodeResponse{Message: "Ok"}, nil
}

// func (s *Handler) GetSeveral(ctx context.Context, in *pb.GetKey) (*pb.GetSeveralKeys, error) {
// 	counter++
// 	val, err := s.kv.GetNode(in.Key)
// 	if err != nil {
// 		return &pb.GetResponse{Value: ""}, err
// 	}
// 	return &pb.GetResponse{Value: val.Value}, nil
// }
