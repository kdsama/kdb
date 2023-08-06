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

// Package main implements a server for Greeter service.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/kdsama/kdb/client"
	"github.com/kdsama/kdb/consensus"
	pb "github.com/kdsama/kdb/consensus/protodata"
	"github.com/kdsama/kdb/logger"
	"github.com/kdsama/kdb/server/store"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	name := os.Args[1]

	leader := false
	filepath := "serverInfo/servers.txt"

	if len(os.Args) > 2 {
		fmt.Println("LEADER ")
		leader = true
	}
	if name == "" {
		log.Fatal("container name is required, exitting")
	}

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	opts := logger.ToOutput(os.Stdout)
	logger := logger.New(0, opts)

	cs := consensus.NewConsensusService(leader, name, filepath, logger)

	go cs.Init()

	kv := store.NewKVService(name, "wal", "../data", 1, 10, logger)

	if leader {
		clientService := client.New(kv, cs, logger)

		go clientService.Automate()
	}

	pb.RegisterConsensusServer(s, &consensus.Receiver{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
