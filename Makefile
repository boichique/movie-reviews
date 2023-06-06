before-push:
	go mod tidy &&\
	gofumpt -l -w . &&\
	go build ./...&&\
	go test -v ./integration_tests/...