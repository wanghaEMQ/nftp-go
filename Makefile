all:
	mkdir -p build && go build -o build/client client.go && go build -o build/server server.go
