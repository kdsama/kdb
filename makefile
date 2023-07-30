run: down up 
new: down upnew


build:
	docker build -t kdb -f "$$(pwd)/deploys/Dockerfile.production" --no-cache .
	
stop:
	docker-compose -f docker-compose.yml stop
down:
	docker-compose -f docker-compose.yml down

test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down 
add:
	go run orchestrator/main.go add kdb
	
delete:
	go run orchestrator/main.go delete kdb

deleteAll:
	go run orchestrator/main.go deleteAll kdb

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative consensus/protodata/consensus.proto