install:
	go install github.com/battlesnakeio/engine/cmd/engine
.PHONY: install

run: install
	engine server
.PHONY: run

run-game: install
	$(eval GAME_ID := $(shell engine create -c ~/snake-config.json | jq '.ID'))
	engine run -g $(GAME_ID)
.PHONY: run-game

test:
	go test -timeout 20s -race -coverprofile coverage.txt -covermode=atomic ./...
.PHONY: test

test-e2e: install
	go test -timeout 120s -race ./e2e -enable-e2e
.PHONY: test-e2e

proto:
	docker run -it --rm -v $$PWD/controller/pb:/build/pb sendwithus/protoc \
		-I /build/pb --gogo_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,plugins=grpc:/build/pb /build/pb/controller.proto
.PHONY: proto

build-docker:
	docker build -t battlesnakeio/engine .
.PHONY: build-docker
