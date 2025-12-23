
GO_BUILD_ENV :=
GO_BUILD_FLAGS :=
MODULE_BINARY := bin/stemma

$(MODULE_BINARY): Makefile go.mod **/*.go cmd/module/*.go 
	$(GO_BUILD_ENV) go build $(GO_BUILD_FLAGS) -o $(MODULE_BINARY) cmd/module/main.go

lint:
	gofmt -s -w .

update:
	go get go.viam.com/rdk@latest
	go mod tidy

test:
	go test ./...

bin/module.tar.gz: meta.json $(MODULE_BINARY)
	strip $(MODULE_BINARY)
	tar czf $@ meta.json $(MODULE_BINARY)

module: test bin/module.tar.gz

all: test bin/module.tar.gz

setup:
	go mod tidy
