GOPATH:=$(shell go env GOPATH)

.PHONE: models
models:
	go generate ./ent

.PHONE: run
run:
	go run ./main.go

.PHONY: deps
deps:
	go mod tidy

.PHONY: swag
swag:
	swag init -g main.go --output docs --parseDependency --parseInternal
build:
	go build -o app

docker:
	docker compose up -d --build --force-recreate