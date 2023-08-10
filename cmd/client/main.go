package main

import (
	"flag"
	"log"
	"net/http"

	clientdiscovery "github.com/kdsama/kdb/clientDiscovery"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	flag.Parse()

	//addServer
	// getServerStatus
	// getkey
	// setkey
	// automate with options --> all go to the handler file
	// setLeader For now we will do it for the first server that we initiate
	// this will be done auto when we receive a request for addServer
	// we need to have rpc connection as well with the server
	// for now I will just use the one in consensus
	// add more methods the consensus one <- consensus is to be used for connect to a client for Reading purposes.
	// we would already be connected to all the clients, just that redirection will be a bit different
	clh := clientdiscovery.NewClientHandler()
	http.HandleFunc("/add-server", clh.AddServer)
	http.HandleFunc("/get", clh.Get)
	http.HandleFunc("/set", clh.Set)
	http.HandleFunc("/automate-get", clh.AutomateGet)
	http.HandleFunc("/automate-set", clh.AutomateSet)

	log.Fatal(http.ListenAndServe(":8080", nil))

}
