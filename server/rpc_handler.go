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
	"os"
	"time"

	"github.com/kdsama/kdb/config"
	pb "github.com/kdsama/kdb/protodata"
	"github.com/prometheus/client_golang/prometheus"
)

type Handler struct {
	pb.UnimplementedConsensusServer
	server *Server
}

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: os.Args[1] + "_ps_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"reqtype"},
	)
	requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    os.Args[1] + "_ps_request_latency",
		Help:    "btree inserts :: btree layer",
		Buckets: []float64{0.0, 20.0, 40.0, 60.0, 80.0, 100.0, 160.0, 180.0, 200.0, 400.0, 800.0, 1600.0},
	}, []string{"reqtype"})
)

func NewHandler(server *Server) *Handler {
	return &Handler{server: server}
}

// Acknowledgement that the heartbeat has been received
func (s *Handler) Ack(ctx context.Context, in *pb.Hearbeat) (*pb.HearbeatResponse, error) {

	// log.Printf("Received: %v", in.GetMessage())
	s.server.HeartbeatAck()
	return &pb.HearbeatResponse{Message: ""}, nil
}

// Record received, now commit/ acknowledge according to the type of data
func (s *Handler) SendRecord(ctx context.Context, in *pb.WalEntry) (*pb.WalResponse, error) {

	switch in.Status {
	case int32(config.Acknowledge):
		// a function is required to just add a wal entry
		t := time.Now()
		s.server.AcknowledgeRecord(&in.Entry)
		requestLatency.WithLabelValues("Acknowledge").Observe(float64(time.Since(t)) / 1000)

	case int32(config.Commit):
		t := time.Now()
		s.server.SetRecord(&in.Entry)
		requestLatency.WithLabelValues("Commit").Observe(float64(time.Since(t)) / 1000)
	}

	return &pb.WalResponse{Message: "ok"}, nil

}
func (s *Handler) Get(ctx context.Context, in *pb.GetKey) (*pb.GetResponse, error) {
	t := time.Now()
	val, err := s.server.Get(in.Key)
	if err != nil {
		return &pb.GetResponse{Value: ""}, err
	}
	requestLatency.WithLabelValues("Get").Observe(float64(time.Since(t)) / 1000)
	return &pb.GetResponse{Value: val}, nil
}
func (s *Handler) Set(ctx context.Context, in *pb.SetKey) (*pb.SetKeyResponse, error) {
	t := time.Now()
	s.server.logger.Infof("Key %v and value %v", in.Key, in.Value)
	err := s.server.Add(in.Key, in.Value)
	if err != nil {
		return &pb.SetKeyResponse{Message: ""}, err
	}
	requestLatency.WithLabelValues("Set").Observe(float64(time.Since(t)) / 1000)
	return &pb.SetKeyResponse{Message: "OK"}, nil
}

func (s *Handler) Broadcast(ctx context.Context, in *pb.BroadcastNode) (*pb.BroadcastNodeResponse, error) {

	// ADD NODE TO EXISTING NODES
	t := time.Now()
	s.server.Broadcast(in.Addr, in.Leader)
	requestLatency.WithLabelValues("Set").Observe(float64(time.Since(t)) / 1000)
	return &pb.BroadcastNodeResponse{Message: "Ok"}, nil
}

func (s *Handler) Vote(ctx context.Context, in *pb.VoteNode) (*pb.VoteNodeResponse, error) {

	// ADD NODE TO EXISTING NODES
	t := time.Now()
	leader, status := s.server.Vote(int(in.Term), in.Leader)
	requestLatency.WithLabelValues("Vote_For_Leader").Observe(float64(time.Since(t)) / 1000)
	return &pb.VoteNodeResponse{Leader: leader, Status: status}, nil
}

func (s *Handler) LeaderInfo(ctx context.Context, in *pb.AskLeader) (*pb.LeaderInfoResponse, error) {
	leader, err := s.server.LeaderInfo()

	return &pb.LeaderInfoResponse{Leader: leader}, err

}

// func (s *Handler) GetSeveral(ctx context.Context, in *pb.GetKey) (*pb.GetSeveralKeys, error) {
// 	counter++
// 	val, err := s.kv.GetNode(in.Key)
// 	if err != nil {
// 		return &pb.GetResponse{Value: ""}, err
// 	}
// 	return &pb.GetResponse{Value: val.Value}, nil
// }

func init() {
	prometheus.MustRegister(requestsTotal, requestLatency)
}
