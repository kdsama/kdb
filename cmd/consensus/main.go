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
	"net/http"
	"os"

	"github.com/kdsama/kdb/consensus"
	"github.com/kdsama/kdb/logger"
	pb "github.com/kdsama/kdb/protodata"
	"github.com/kdsama/kdb/store"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	name := os.Args[1]

	if name == "" {
		log.Fatal("container name is required, exitting")
	}

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	opts := logger.ToOutput(os.Stdout)
	logger := logger.New(logger.Info, opts)

	cs := consensus.NewConsensusService(name, logger)

	go cs.Init()

	kv := store.NewKVService("./data/kvservice/persist/", "node", "./data/kvservice/wal/", 1, 1000, logger)

	go ServerHttp()
	pb.RegisterConsensusServer(s, consensus.NewHandler(kv))

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
func ping(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "pong")
}
func ServerHttp() {
	http.HandleFunc("/ping", ping)
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":8080", nil))

}
