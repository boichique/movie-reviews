before-push:
	go mod tidy &&\
	gofumpt -l -w . &&\
	go build ./...&&\
	golangci-lint run ./... &&\
	go test -v ./tests/...

scrap:
	go run ./scrapper scrap -o ./scrapper/output
ingest:
	go run ./scrapper ingest -i ./scrapper/output --email admin@movies.com --password "paSsw0rd!"
