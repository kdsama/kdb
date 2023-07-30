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

func Run(addr string) {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewConsensusClient(conn)

	// Contact the server and print out its response.
	i := 0
	for {
		i++
		ctx, _ := context.WithTimeout(context.Background(), time.Second)

		r, err := c.Ack(ctx, &pb.Hearbeat{Message: (fmt.Sprint(i))})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		time.Sleep(10 * time.Second)

		log.Printf("Greeting: from %s %s", addr, r.GetMessage())
	}

}
