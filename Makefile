BINARY_DIR = bin
BINARIES = $(patsubst cmd/%/,%,$(wildcard cmd/*/))

.PHONY: all build test validate clean run $(BINARIES)

all: build
	

build: $(BINARIES)
	

$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

$(BINARIES): %: $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$@ ./cmd/$@/

test:
	go test ./...

validate:
	@test -z "$(shell gofumpt -l .)" && echo "No files need formatting" || (echo "Incorrect formatting in:"; gofumpt  -l .; exit 1)
	go vet ./...

clean:
	rm -rf $(BINARY_DIR)
	go clean

run: $(LINKSERV)
	$(LINKSERV)
