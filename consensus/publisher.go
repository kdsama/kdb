package consensus

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/kdsama/kdb/consensus/protodata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Run(node string, nc *Client, leader bool) {
	// Set up a connection to the server.
	c := (*nc).con
	// Contact the server and print out its response.
	if leader {
		fmt.Println("Hi I am leader")
		i := 0
		for {
			i++

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

			r, err := (*c).Ack(ctx, &pb.Hearbeat{Message: (fmt.Sprint(i))})
			if err != nil {
				log.Fatalf("could not greet: %v", err)
			}
			time.Sleep(2 * time.Second)
			log.Printf("Greeting: from %s %s", node, r.GetMessage())
		}
	}

}

func connect(addr string) pb.ConsensusClient {
	flag.Parse()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewConsensusClient(conn)
	return c
}
