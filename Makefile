TARGET   ?= osmedeus
GO       ?= go
GOFLAGS  ?= 
VERSION  := $(shell cat libs/version.go | grep 'VERSION =' | cut -d '"' -f 2)

build:
	go install
	go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/osmedeus

release:
	go install
	# this is only for local build
	rm -rf ./dist/* ~/myGit/premium-osmedeus-base/dist/* ~/org-osmedeus/osmedeus-base/dist/*
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/osmedeus
	zip -9 -j dist/osmedeus-macos.zip dist/osmedeus
	rm -rf ./dist/osmedeus
	# for linux build on mac
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -tags netgo -trimpath -buildmode=pie -o dist/osmedeus
	zip -j dist/osmedeus-linux.zip dist/osmedeus
	rm -rf ./dist/osmedeus
	cp dist/* ~/myGit/premium-osmedeus-base/dist/
	cp dist/* ~/org-osmedeus/osmedeus-base/dist/
	$(TARGET) update --gen dist/public.json
	mv dist/osmedeus-macos.zip dist/osmedeus-$(VERSION)-macos.zip
	mv dist/osmedeus-linux.zip dist/osmedeus-$(VERSION)-linux.zip
run:
	$(GO) $(GOFLAGS) run *.go

fmt:
	$(GO) $(GOFLAGS) fmt ./...; \
	echo "Done."

test:
	$(GO) $(GOFLAGS) test ./... -v%