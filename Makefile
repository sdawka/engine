install:
	go install

run: install
	engine

test:
	go test ./...