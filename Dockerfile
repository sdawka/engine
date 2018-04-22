FROM golang:1.10.1-alpine as builder
WORKDIR /go/src/github.com/battlesnakeio/engine/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go install -installsuffix cgo ./cmd/...
RUN CGO_ENABLED=0 GOOS=linux go install -installsuffix cgo .

FROM alpine:latest as certs
RUN apk add --no-cache ca-certificates

FROM scratch
ENV PATH=/bin
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/bin/ /bin/
CMD ["/bin/engine"]
