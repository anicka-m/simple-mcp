BINARY_NAME=simple-mcp
GO=go
LDFLAGS=-ldflags="-w -s"
BUILD_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64

build:
	$(BUILD_ENV) $(GO) build $(LDFLAGS) -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)
