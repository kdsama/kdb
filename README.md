# kdb
Distributed Key-Value Store (Master's Project)


# Distributed Key-Value Store

<!--- ![Project Demo](demo.gif) <!-- Replace with a link to your project demo or a GIF showcasing it --> 

## Overview

This project showcases a distributed key-value store implementation in Golang. The key-value store allows clients to store, retrieve, and delete key-value pairs in a distributed environment, highlighting proficiency in Golang and distributed systems concepts.

## Features

- **Distributed Architecture:** The project utilizes a distributed architecture to ensure scalability and fault tolerance in the key-value store.

- **Consistent Hashing:** Consistent hashing is employed to distribute data across nodes uniformly, preventing hotspots and ensuring efficient data retrieval.

- **Data Replication:** Key-value pairs are replicated across multiple nodes to enhance data availability and reliability.

- **CRUD Operations:** The key-value store supports Create, Read, Update, and Delete operations through an intuitive API.

- **Concurrency Management:** Concurrency control mechanisms are implemented to handle concurrent read and write requests seamlessly.

- **Failure Handling:** The system gracefully handles node failures, reallocating data and maintaining data integrity.

## Technologies and Concepts Applied

- Golang programming language
- Distributed systems principles
- Networking libraries (e.g., `net` package)
- Consistent hashing algorithm
- Data replication strategies
- Concurrency control mechanisms

## Folder Structure

- `client`: Contains the client application for interacting with the distributed key-value store.
- `node`: Holds the code for the individual nodes in the distributed system.
- `store`: Includes the implementation of the distributed key-value storage logic.
- `utils`: Houses utility functions and shared components used across the project.

## Getting Started

1. Clone the repository: `git clone https://github.com/kdsama/kdb.git`
2. Navigate to the project directory: `cd kdb`
3. Follow the setup instructions in the respective subfolders (`client`, `node`, `store`) to configure and run the components.

## Usage

1. Start the nodes: Navigate to the `node` folder and run `go run node.go --port <port_number>` for each node.
2. Interact with the key-value store using the client application located in the `client` folder. Run the client with `go run client.go --server <server_address>`.

## Run
You can utilize the makefile. Here are the commands 
To run the project, you can utilize the provided Makefile. Here are some useful commands:

- `make run`: Stops any running containers, builds the Docker image, and brings up the project containers.
- `make new`: Similar to `make run`, but also cleans the previous data before starting.
- `make build`: Builds the Docker image for the key-value store.
- `make stop`: Stops the project containers.
- `make down`: Stops and removes the project containers.
- `make test`: Runs tests using the test configuration.
- `make add`: Adds a key-value pair using the client. Adds 5 nodes. 
- `make delete`: Deletes a key-value pair using the client.
- `make deleteAll`: Deletes all key-value pairs using the client.
- `make proto`: Generates Go code from the consensus.proto file.
- `make reload`: Deletes all data, rebuilds the Docker image, adds data, and restarts Prometheus. 
- `make restart`: Deletes all data, rebuilds the Docker image, and adds data.


## Contribution

Contributions to this project are welcome! To contribute:
1. Fork the repository
2. Create a new branch
3. Commit your changes
4. Push the branch and open a pull request

## License

This project is licensed under the [MIT License](LICENSE).

## Contact

For any inquiries or suggestions, contact [kshitij.dhingra5@gmail.com].

---
Customize this template with your specific project details, setup instructions, and usage guidelines.


Please integrate your code and details from the previous messages into this template to create a comprehensive `README.md` file for your distributed key-value store project.
