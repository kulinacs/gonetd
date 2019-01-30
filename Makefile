.PHONY: vet test cover

test: vet
	go test -v -cover ./...

cover: vet
	go test -v -coverprofile=coverage.out -timeout 10s ./...
	go tool cover -html=coverage.out

vet:
	go vet ./...
