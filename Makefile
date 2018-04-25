install:
	go install -v

run: install
	engine

make run-game: install-cli
	$(eval GAME_ID := $(shell engine-cli create -c ~/snake-config.json | jq '.ID'))
	engine-cli run -g $(GAME_ID)

install-cli:
	go install github.com/battlesnakeio/engine/cmd/engine-cli

test:
	go test -timeout 20s -race -coverprofile coverage.txt -covermode=atomic ./...

proto:
	docker run -it --rm -v $$PWD/controller/pb:/build/pb sendwithus/protoc \
		-I /build/pb --gogo_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,plugins=grpc:/build/pb /build/pb/controller.proto

build-docker:
	docker build -t battlesnakeio/engine .