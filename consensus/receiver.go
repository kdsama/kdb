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
	"fmt"
	"time"

	pb "github.com/kdsama/kdb/consensus/protodata"
	"github.com/kdsama/kdb/server/store"
)

type Receiver struct {
	pb.UnimplementedConsensusServer
	kv *store.KVService
}

func NewReciever(kv *store.KVService) *Receiver {
	return &Receiver{kv: kv}
}

// Acknowledgement that the heartbeat has been received
func (s *Receiver) Ack(ctx context.Context, in *pb.Hearbeat) (*pb.HearbeatResponse, error) {

	// log.Printf("Received: %v", in.GetMessage())
	return &pb.HearbeatResponse{Message: "Hello " + in.GetMessage()}, nil
}

var counter = 0

// Record received, now commit/ acknowledge according to the type of data
func (s *Receiver) SendRecord(ctx context.Context, in *pb.WalEntry) (*pb.WalResponse, error) {
	counter++

	switch in.Status {
	case int32(Acknowledge):
		// a function is required to just add a wal entry

		s.kv.AcknowledgeRecord(&in.Entry)

	case int32(Commit):
		fmt.Println(counter, "Receiving Commit stuff at ", time.Now())
		s.kv.SetRecord(&in.Entry)
	}

	return &pb.WalResponse{Message: "ok"}, nil
}
