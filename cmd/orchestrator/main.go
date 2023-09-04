package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kdsama/kdb/cmd/orchestrator/prom"
)

type dockercli struct {
	*client.Client
	image string
}

const (
	IMAGE   = "kdb"
	NETWORK = "kdb_backend"
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

	if len(args) > 4 {
		log.Fatal("Exiting, cannot take more than 1 arguments")
	}
	switch args[1] {
	case "list":
		dc.listContainers()
	case "add":
		dc.addContainers(args[3])
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

func (dc *dockercli) addContainers(times string) {
	dc.addContainer(times)
}

func (dc *dockercli) addContainer(times string) {
	n, _ := strconv.Atoi(times)
	ps := prom.NewPrometheusConfig()
	ps.ResetData()
	count := 0
	for i := 0; i < n; i++ {

		// check if kdb_backend exists
		nw, err := dc.NetworkList(context.Background(), types.NetworkListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, network := range nw {
			if network.Name == NETWORK {
				count++
			}
		}
		if count == 0 {
			dc.NetworkCreate(context.Background(), NETWORK, types.NetworkCreate{})
		}

		rand.Seed(time.Now().UnixNano())
		randId := fmt.Sprintf("%v", rand.Int31n(100000))
		name := "node" + randId
		rand.Seed(time.Now().UnixNano())
		name2 := "node" + fmt.Sprintf("%v", rand.Int31n(100000))

		var exposedPorts nat.PortSet
		portBindings := nat.PortMap{
			"8080/tcp": []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: "8080"}},
		}

		volume, _ := filepath.Abs(filepath.Dir("data/"))

		volume += "/" + "data" + randId
		err1 := os.Mkdir(volume, 0755)
		if err1 != nil {
			log.Fatal(err1)
		}

		// write to server info file
		cmd := []string{"./bin/serve", "-name", name}
		// check if containers existed previously
		arr := dc.listContainers()
		if len(arr) == 0 {

			exposedPorts = nat.PortSet{"8080/tcp": {}}
			resp, err := dc.ContainerCreate(context.Background(), &container.Config{
				Image:        dc.image,
				Cmd:          []string{"./bin/serveClient", name2},
				Env:          []string{"GRPC_GO_LOG_VERBOSITY_LEVEL=99", "GRPC_GO_LOG_SEVERITY_LEVEL=info "},
				ExposedPorts: exposedPorts,
			}, &container.HostConfig{PortBindings: portBindings}, // Binds: []string{volume + ":/go/src/data"}
				&network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{NETWORK: {NetworkID: NETWORK}}}, nil, name2)

			if err != nil {
				panic(err)
			}
			if err := dc.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
				panic(err)
			}

			ps.AddScrapeConfig(name2, ":8080")
		}
		rand.Seed(time.Now().UnixNano())
		prt := 50051 + rand.Intn(1000)
		prmPort := fmt.Sprintf(":%d", 8080+rand.Intn(10000))
		cmd = append(cmd, "-port", fmt.Sprint(prt), "-promport", prmPort)
		fmt.Println("We are goint o add port ", prt, cmd)
		resp, err := dc.ContainerCreate(context.Background(), &container.Config{
			Image:        dc.image,
			Cmd:          cmd,
			ExposedPorts: exposedPorts,
		}, &container.HostConfig{
			Binds: []string{volume + ":/go/src/data"}},
			&network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{NETWORK: {NetworkID: NETWORK}}}, nil, name)

		if err != nil {
			panic(err)
		}
		if err := dc.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}
		ps.AddScrapeConfig(name, prmPort)
		time.Sleep(1 * time.Second)
		http.Get("http://localhost:8080/add-server?name=" + name + fmt.Sprintf(":%d", prt))

	}
	ps.Generate()
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
func (dc *dockercli) listContainers() []string {

	containers, err := dc.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {

		panic(err)
	}
	arr := []string{}

	for _, container := range containers {

		if container.Image == dc.image {

			arr = append(arr, strings.Split(container.Names[0], "/")[1])
		}
	}
	return arr
}
