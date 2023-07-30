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