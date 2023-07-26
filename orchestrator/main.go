package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type dockercli struct {
	*client.Client
	image string
}

const (
	IMAGE   = "go-docker-grpc_server"
	NETWORK = "go-docker-grpc_server_backend"
)

func main() {

	// I had to go inside the package and change the client version for the go-docker api
	args := os.Args

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	dc := dockercli{cli, args[2]}
	defer cli.Close()

	if len(args) > 3 {
		log.Fatal("Exiting, cannot take more than 1 arguments")
	}
	switch args[1] {
	case "list":
		dc.listContainers()
	case "add":
		dc.addContainer()
	case "delete":
		if len(args) == 4 {
			dc.delete(args[3])
		} else {
			dc.delete("")
		}

	case "deleteAll":
		dc.deleteAll()
	}
}

func (dc *dockercli) addContainer() {
	resp, err := dc.ContainerCreate(context.Background(), &container.Config{
		Image: dc.image,
	}, nil, nil, nil, "")
	fmt.Println(resp)
	if err != nil {
		panic(err)
	}
	if err := dc.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

}

func (dc *dockercli) delete(containerID string) {
	toDelete := containerID
	if strings.Trim(containerID, " ") == "" {
		// delete a random one
		containers, err := dc.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		arr := []string{}
		for _, con := range containers {
			if con.Image == dc.image {
				arr = append(arr, con.ID)

			}
		}
		toDelete = arr[rand.Intn(len(arr))]
	}
	err := dc.ContainerStop(context.Background(), toDelete, container.StopOptions{})
	if err != nil {
		log.Fatal(err)
	}
	err = dc.ContainerRemove(context.Background(), toDelete, types.ContainerRemoveOptions{})
	if err != nil {
		log.Fatal(err)
	}
}

func (dc *dockercli) deleteAll() {
	containers, err := dc.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {

		panic(err)
	}

	for _, con := range containers {

		if con.Image == dc.image {
			err := dc.ContainerStop(context.Background(), con.ID, container.StopOptions{})
			if err != nil {
				log.Fatal(err)
			}
			err = dc.ContainerRemove(context.Background(), con.ID, types.ContainerRemoveOptions{})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
func (dc *dockercli) listContainers() {

	containers, err := dc.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {

		panic(err)
	}

	for _, container := range containers {

		if container.Image == dc.image {
			fmt.Println(strings.Split(container.Names[0], "/")[1])
		}
	}
}
