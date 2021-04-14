
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
SRV_TARGET=goconnect/cmd/server 
CLI_TARGET=goconnect/cmd/client 
CLI_LIB=goconnect/cmd/client

SRV_BINARY_NAME=sg_server
CLI_BINARY_NAME=sg_client
CLI_LIB_NAME=sg_client.so

all: build build-lib
build: 
	$(GOBUILD) -o $(SRV_BINARY_NAME) -v $(SRV_TARGET)
	$(GOBUILD) -o $(CLI_BINARY_NAME) -v $(CLI_TARGET)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(CLI_BINARY_NAME).exe -v $(CLI_TARGET)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(SRV_BINARY_NAME).exe -v $(SRV_TARGET)

build-lib:
	$(GOBUILD) -o $(CLI_LIB_NAME) -buildmode=c-shared gsclient.go

test:
	$(GOTEST) -v ./...

clean: 
	$(GOCLEAN)
	rm -f $(SRV_BINARY_NAME) $(SRV_BINARY_NAME).exe
	rm -f $(CLI_BINARY_NAME) $(CLI_BINARY_NAME).exe
	rm -f $(CLI_LIB_NAME)

runs: build
	./$(SRV_BINARY_NAME)
