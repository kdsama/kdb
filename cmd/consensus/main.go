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
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/kdsama/kdb/config"
	"github.com/kdsama/kdb/consensus"
	pb "github.com/kdsama/kdb/protodata"
	"github.com/kdsama/kdb/server"
	"github.com/kdsama/kdb/store"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var cpuprofile = flag.String("cpuprofile", "./data/cpu.prof", "write cpu profile to file")
var memprofile = flag.String("memprofile", "./data/mem.prof", "write cpu profile to file")

func main() {

	var (
		prt  = flag.Int("port", 50051, "The server port")
		pprt = flag.String("promport", ":8080", "The server port")
		nm   = flag.String("name", "", "This is the name of the node")
	)
	flag.Parse()

	var (
		port     = *prt
		name     = *nm
		promPort = *pprt
		lis, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		s        *grpc.Server
		// opts     = logger.ToOutput(os.Stdout)
		// opts1    = logger.DateOpts(false)
		logger = setupLogger()
	)
	if *cpuprofile != "" {
		fmt.Println("Running with CPU profile")
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("Finished")
		if *memprofile != "" {
			f, err := os.Create(*memprofile)
			if err != nil {
				log.Fatal(err)
			}
			f1, err := os.Create(*memprofile + "1")
			if err != nil {
				log.Fatal(err)
			}
			f2, err := os.Create(*memprofile + "2")
			if err != nil {
				log.Fatal(err)
			}
			f3, err := os.Create(*memprofile + "3")
			if err != nil {
				log.Fatal(err)
			}
			runtime.GC()
			pprof.Lookup("heap").WriteTo(f, 0)
			pprof.Lookup("goroutine").WriteTo(f1, 0)
			pprof.Lookup("threadcreate").WriteTo(f2, 0)
			pprof.Lookup("block").WriteTo(f3, 0)
			defer f.Close()
		}
		if *cpuprofile != "" {
			pprof.StopCPUProfile()
		}

		os.Exit(0)
	}()
	if name == "" {
		log.Fatal("container name is required, exitting")
	}

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s = grpc.NewServer()

	cs := consensus.NewConsensusService(fmt.Sprintf("%s:%d", name, port), logger)

	go cs.Init()

	// initializing services
	kv := store.NewKVService(config.DataPrefix, config.WalPrefix, config.Directory, config.WalBufferInterval, config.BtreeDegree, logger)
	kv.Init()
	SR := server.New(kv, cs, logger)

	go ServerHttp(promPort)

	pb.RegisterConsensusServer(s, server.NewHandler(SR))
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	// <-sig

}

// Serves Http Requests. Open for Prometheus to fetch metrics
func ServerHttp(promPort string) {
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal("????????????????????", http.ListenAndServe(promPort, nil))

}

func setupLogger() *zap.SugaredLogger {

	l, _ := zap.NewProduction()
	return l.Sugar()
}
