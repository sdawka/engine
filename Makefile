install:
	go install github.com/battlesnakeio/engine/cmd/engine
.PHONY: install

run: install
	engine
.PHONY: run

make run-game: install-cli
	$(eval GAME_ID := $(shell engine-cli create -c ~/snake-config.json | jq '.ID'))
	engine-cli run -g $(GAME_ID)
.PHONY: run-game

install-cli:
	go install github.com/battlesnakeio/engine/cmd/engine-cli
.PHONY: install-cli

test:
	go test -timeout 20s -race -coverprofile coverage.txt -covermode=atomic ./...
.PHONY: test

test-e2e: install
	go test -timeout 120s -race ./e2e -enable-e2e
.PHONY: test-e2e

lint:
	gometalinter --config ./.gometalinter.json ./...
.PHONY: lint

proto:
	docker run -it --rm -v $$PWD/controller/pb:/build/pb sendwithus/protoc \
		-I /build/pb --gogo_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,plugins=grpc:/build/pb /build/pb/controller.proto
.PHONY: proto

build-docker:
	docker build -t battlesnakeio/engine .
.PHONY: build-docker
