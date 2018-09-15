GOCMD := go
GOTEST := $(GOCMD) test -v -count=1
GOBUILD := $(GOCMD) build

PKGS := $(shell go list ./... | grep -v vendor)
MAINPKGS := $(foreach cmd_dir, $(shell go list ./cmd/...), $(basename $(cmd_dir)))

.PHONY: test all
all: $(MAINPKGS)


.PHONY: test
test:
	$(GOTEST) $(PKGS)


$(MAINPKGS):
	$(GOBUILD) -o bin/$(@F) $@

.PHONY: clean
clean:
	@rm -rf bin

