#!/bin/bash

docker run -it --rm -v \
  $PWD:/build/pb \
  sendwithus/protoc \
  -I /build/pb \
  --gogo_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,plugins=grpc:/build/pb /build/pb/controller.proto
