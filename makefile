run: down up 
new: down upnew


up:
	docker build -t kdb -f "$$(pwd)/deploys/Dockerfile.production" .
	docker run -it -d -p 5869:5869 --name kdb_node_leader --network backend -v ./data:/go/src/data
stop:
	docker-compose -f docker-compose.yml stop
down:
	docker-compose -f docker-compose.yml down

test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down 
add:
	docker run -d --network=go-docker-grpc_backend go-docker-grpc_server 
delete:
	go run orchestrator/main.go delete go-docker-grpc_server