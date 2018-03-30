FROM golang:1.10-alpine
WORKDIR /go/src/github.com/battlesnakeio/engine/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go install -installsuffix cgo ./cmd/...

FROM alpine:3.7
RUN apk add --no-cache ca-certificates
COPY --from=0 /go/bin/ /bin/
