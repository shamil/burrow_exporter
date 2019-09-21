# The crossbuild command sets the PREFIX env var for each platform
PREFIX ?= $(shell pwd)

.PHONY: build
build:
	@command -v promu >/dev/null || { \
		echo ">> installing promu"; \
		GO111MODULE=off GOOS= GOARCH= go get -u github.com/prometheus/promu; \
	}

	@promu build --prefix $(PREFIX)

.PHONY: crossbuild
crossbuild:
	@promu crossbuild
	@promu crossbuild tarballs

.PHONY: release
release: crossbuild
	# this should be invoked after drafting the release in github.com
	@promu release .tarballs
