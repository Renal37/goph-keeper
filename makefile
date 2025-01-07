.PHONY: build-server build-agent cert

build:
	make cert
	make build-server
	make build-agent

build-server:
	cd cmd/server; GOOS=linux GOARCH=amd64 go build -o server-linux-amd64 main.go; cd ../..
	cd cmd/server; GOOS=windows GOARCH=amd64 go build -o server-win-amd64 main.go; cd ../..
	cd cmd/server; GOOS=darwin GOARCH=amd64 go build -o server-darwin-amd64 main.go; cd ../..

build-agent:
	cd cmd/agent; GOOS=linux GOARCH=amd64 go build -o agent-linux-amd64 main.go; cd ../..
	cd cmd/agent; GOOS=windows GOARCH=amd64 go build -o agent-win-amd64 main.go; cd ../..
	cd cmd/agent; GOOS=darwin GOARCH=amd64 go build -o agent-darwin-amd64 main.go; cd ../..

cert:
	cd cert; ./gencert.sh; cd ..
