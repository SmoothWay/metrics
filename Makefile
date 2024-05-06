###
### Makefile
###

VERSION=0.0.1dev

B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags)
REV=$(GITREV)-$(BRANCH)-$(shell date +%Y%m%d-%H:%M:%S)
DATE=$(shell date +%Y%m%d-%H:%M:%S)
COMMIT=$(shell git log -n 1 --pretty=format:"%H")

info:
	- @echo "revision $(REV)"

build: info
	@ echo
	@ echo "Compiling Binary"
	@ echo
	cd cmd/server && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildCommit=$(COMMIT) -X main.buildDate=$(DATE) -s -w" -o server
	cd cmd/agent && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildCommit=$(COMMIT) -X main.buildDate=$(DATE) -s -w" -o agent

build_macos: info
	@ echo
	@ echo "Compiling Binary for MacOS"
	@ echo
	cd cmd/server && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildCommit=$(COMMIT) -X main.buildDate=$(DATE) -s -w" -o server
	cd cmd/agent && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildCommit=$(COMMIT) -X main.buildDate=$(DATE) -s -w" -o agent

tidy:
	@ echo
	@ echo "Tidying"
	@ echo
	go mod tidy

clean:
	@ echo
	@ echo "Cleaning"
	@ echo
	rm cmd/server/server
	rm cmd/agent/agent

utest: build
	@ echo
	@ echo "Unit testing"
	@ echo
	go test ./...

test: build
	@ echo
	@ echo "Testing"
	@ echo
	metricstest -test.v -test.run=^TestIteration1\$$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -source-path=. -server-port 8080

run:
	@echo "Running server"
	@go run ./cmd/server/ &
	@echo "Running agent"
	@go run ./cmd/agent/

PHONY: build tidy clean utest test run