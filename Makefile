SHELL=/usr/bin/env bash

all: ffi build

unexport GOFLAGS

ldflags=-X=github.com/OpenFilWallet/OpenFilWallet/build.CurrentCommit=+git.$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))
ifneq ($(strip $(LDFLAGS)),)
 ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"

.PHONY: ffi
ffi:
	git submodule update --init --recursive
	./extern/filecoin-ffi/install-filcrypto

.PHONY: build
build: build
	go mod tidy -go=1.16 && go mod tidy -go=1.17
	rm -rf openfild
	go build $(GOFLAGS) -o openfild ./cmd/openfild

