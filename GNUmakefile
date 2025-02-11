BIN	= scanapi

GO	:= go

# https://github.com/golang/go/issues/64875
arch := $(shell uname -m)
ifeq ($(arch),s390x)
CGO_ENABLED = 0
else
CGO_ENABLED := 1
endif

all:	$(BIN)

$(BIN): *.go
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -trimpath -ldflags="-s -w -buildid=" -buildmode=pie

.PHONY: test
test:
	@$(GO) vet
	@staticcheck

.PHONY: clean
clean:
	@$(GO) clean

.PHONY: gen
gen:
	@rm -f go.mod go.sum
	@$(GO) mod init $(BIN)
	@$(GO) mod tidy
