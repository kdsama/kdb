run: down up 
new: down upnew


build:
	docker build -t kdb -f "$$(pwd)/deploys/Dockerfile.production" .
	
stop:
	docker-compose -f docker-compose.yml stop
down:
	docker-compose -f docker-compose.yml down

test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down 
add:
	go run cmd/orchestrator/main.go add kdb 5
	
delete:
	go run cmd/orchestrator/main.go delete kdb

deleteAll:
	go run cmd/orchestrator/main.go deleteAll kdb

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative protodata/consensus.proto

reload:
	make deleteAll
	sudo rm -rf data/*
	make build 
	make add
	docker restart prom 


restart:
	make deleteAll
	sudo rm -rf data/*
	make add



startProm: 
	docker run -p 9090:9090 -d -v $$(pwd)/prometheus/prom.yml:/etc/prometheus/prometheus.yml --network=kdb_backend  --name prom prom/prometheus

startGraf:
	docker run -d -p 3000:3000 --network=kdb_backend --name grafana grafana/grafana

runHere: 
	go run ./cmd/consensus/main.go -name 127.0.0.1 -promport :8081 -port 5051
	go run ./cmd/consensus/main.go -name 127.0.0.1 -promport :8082 -port 5052
	go run ./cmd/consensus/main.go -name 127.0.0.1 -promport :8083 -port 5053
	go run ./cmd/client/main.go -name 127.0.0.1 -promport :8083 -port 8080
