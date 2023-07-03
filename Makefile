PWD:=$(shell pwd)
VERSION=0.0.0
MONOVA:=$(shell which monova dot 2> /dev/null)

version:
ifdef MONOVA
override VERSION=$(shell monova)
else
	$(info "Install monova (https://github.com/jsnjack/monova) to calculate version")
endif

bin/template-renderer: version main.go cmd/*.go
	CGO_ENABLED=0  go build -ldflags="-X github.com/surfly/template-renderer/cmd.Version=${VERSION}" -o bin/template-renderer

test:
	cd cmd && go test

build: test bin/template-renderer

release: build
	grm release surfly/template-renderer -f bin/template-renderer -t "v`monova`"

.ONESHELL:
clean:
	rm -rf bin/*

.PHONY: version release build test
