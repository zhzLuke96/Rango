GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

BINARY_NAME=rango_example

EXAMPLE_FOLDER=./example
SOURCE_FOLDER=./rango

BUILD_CMD=$(GOBUILD) -v -o ${BINARY_NAME}

build-all: build-linux build-windows build-mac

build-linux: deps
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(BUILD_CMD)_unix ${EXAMPLE_FOLDER}

build-windows: deps
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(BUILD_CMD)_windows ${EXAMPLE_FOLDER}

build-mac: deps
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(BUILD_CMD)_darwin ${EXAMPLE_FOLDER}

build:
	BUILD_CMD ${EXAMPLE_FOLDER}

test:
	$(GOTEST) -v $(SOURCE_FOLDER)

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)_unix
	rm -f $(BINARY_NAME)_windows
	rm -f $(BINARY_NAME)_darwin

deps: ;
