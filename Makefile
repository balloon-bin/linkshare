BINARY_DIR = bin
BINARIES = $(patsubst cmd/%/,%,$(wildcard cmd/*/))

.PHONY: all build test validate clean run $(BINARIES)

VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
COMMIT_DATETIME := $(shell git log -1 --format=%cd --date=iso8601)

LDFLAGS := -X git.omicron.one/omicron/linkshare/internal/version.Version=$(VERSION) \
           -X git.omicron.one/omicron/linkshare/internal/version.GitCommit=$(COMMIT) \
           -X "git.omicron.one/omicron/linkshare/internal/version.CommitDateTime=$(COMMIT_DATETIME)"
OPEN = xdg-open


all: build
	

build: $(BINARIES)
	

$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

$(BINARIES): %: $(BINARY_DIR)
	go build -ldflags '$(LDFLAGS)' -o $(BINARY_DIR)/$@ ./cmd/$@/

test:
	mkdir -p reports/coverage/
	go test ./... -coverprofile=reports/coverage/coverage.out
	go tool cover -html=reports/coverage/coverage.out -o reports/coverage/coverage.html && $(OPEN) reports/coverage/coverage.html


validate:
	@test -z "$(shell gofumpt -l .)" && echo "No files need formatting" || (echo "Incorrect formatting in:"; gofumpt  -l .; exit 1)
	go vet ./...

clean:
	rm -rf $(BINARY_DIR)
	rm -rf reports
	go clean

run: $(LINKSERV)
	$(LINKSERV)
