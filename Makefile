install:
	go install -v

run: install
	engine

test:
	go test -coverprofile c.out ./...

proto:
	docker run -it --rm -v \
		$$PWD/controller/pb:/build/pb \
		sendwithus/protoc \
		-I /build/pb \
		--gogo_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,plugins=grpc:/build/pb /build/pb/controller.proto