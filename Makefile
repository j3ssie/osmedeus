TARGET   ?= osmedeus
GO       ?= go
GOFLAGS  ?= 
VERSION  := $(shell cat libs/version.go | grep 'VERSION =' | cut -d '"' -f 2)

build:
	go install
	go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/osmedeus

release:
	go install
	@echo "==> Clean up old builds"
	rm -rf ./dist/* ~/myGit/premium-osmedeus-base/dist/* ~/org-osmedeus/osmedeus-base/dist/*
	@echo "==> building binaries for for mac intel"
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/osmedeus
	zip -9 -j dist/osmedeus-macos-amd64.zip dist/osmedeus && rm -rf ./dist/osmedeus
	@echo "==> building binaries for for mac M1 chip"
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/osmedeus
	zip -9 -j dist/osmedeus-macos-arm64.zip dist/osmedeus&& rm -rf ./dist/osmedeus
	@echo "==> building binaries for linux intel build on mac"
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/osmedeus
	zip -j dist/osmedeus-linux.zip dist/osmedeus&& rm -rf ./dist/osmedeus
	cp dist/* ~/myGit/premium-osmedeus-base/dist/
	cp dist/* ~/org-osmedeus/osmedeus-base/dist/
	@echo "==> Generating metadata info"
	$(TARGET) update --gen dist/public.json
	mv dist/osmedeus-macos-amd64.zip dist/osmedeus-$(VERSION)-macos-amd64.zip
	mv dist/osmedeus-macos-arm64.zip dist/osmedeus-$(VERSION)-macos-arm64.zip
	mv dist/osmedeus-linux.zip dist/osmedeus-$(VERSION)-linux.zip
run:
	$(GO) $(GOFLAGS) run *.go

fmt:
	$(GO) $(GOFLAGS) fmt ./...; \
	echo "Done."

test:
	$(GO) $(GOFLAGS) test ./... -v%