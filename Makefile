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
	swag init -g ./internal/app/app.go -o docs/swagger --parseDepth 10 --parseDependency --parseInternal \
		--outputTypes go,json --parseGoList false --propertyStrategy camelcase --pd true -d ./internal
build:
	go build -o app

docker:
	docker compose up -d --build