package consensus

import (
	"bufio"
	"log"
	"os"
)

// here I need to implement the acknowledgement first
// I need a map of grpc connections as well
// create a map of all the connections for all the servers

func ReadFromFile(filepath string) ([]string, error) {
	t := []string{}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t = append(t, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return t, nil
}

func RunServers(name string) {
	// get the list of address from the servers.txt
	// run the below function in different goroutines
	// lets start with just

	arr, err := ReadFromFile("serverInfo/servers.txt")
	if err != nil {
		log.Fatal(err)
	}
	for _, node := range arr {

		ap := "node" + node
		if ap == name {
			continue
		}
		go Run(ap + ":50051")
	}

}
