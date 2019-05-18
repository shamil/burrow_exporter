# The crossbuild command sets the PREFIX env var for each platform
PREFIX ?= $(shell pwd)

.PHONY: build
build:
	@go build -o $(PREFIX)/burrow_exporter

.PHONY: crossbuild
crossbuild:
	@promu crossbuild
	@promu crossbuild tarballs

.PHONY: release
release: crossbuild
	# this should be invoked after drafting the release in github.com
	@promu release .tarballs
