.PHONY: build build_app
build: build_app
build_app:
	go build -o bin/app cmd/bcmon/main.go
run-local-metrics:
	echo "docker will run on backgroud, pls do `docker ps` for get list"
	docker compose -f ./graph-node/docker-compose.yml  up -d
