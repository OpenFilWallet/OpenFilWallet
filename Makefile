SHELL=/usr/bin/env bash

all: ffi vue build

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
	go mod tidy
	rm -rf openfild
	go build $(GOFLAGS) -o openfild ./cmd/openfild
	go build $(GOFLAGS) -o openfil-cli ./cmd/cli

.PHONY: vue
vue: vue
	cd ./webui && npm install
	cd ./webui && npm run build:prod

install: install
	sudo mv openfild openfil-cli /usr/local/bin/