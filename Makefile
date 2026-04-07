.PHONY: dev down logs run-server run-cli build tidy fmt package-install

dev:
	docker compose -f deploy/dev/docker-compose.yml up --build --force-recreate -d

down:
	docker compose -f deploy/dev/docker-compose.yml down

logs:
	docker compose -f deploy/dev/docker-compose.yml logs -f app

run-server:
	go run ./cmd/server

run-cli:
	go run ./cmd/cli --help

build:
	go build ./...

package-install:
	./scripts/release/package-install.sh

tidy:
	go mod tidy

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')
