before-push:
	go mod tidy &&\
	gofumpt -l -w . &&\
	go build ./...&&\
	golangci-lint run ./... &&\
	go test -v ./integrationTests/...