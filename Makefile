TARGET   ?= osmedeus
GO       ?= go
GOFLAGS  ?= 

build:
	go install
	rm -rf ./dist/*
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o dist/osmedeus
	zip -9 -j dist/osmedeus-macos.zip dist/osmedeus
	rm -rf ./dist/osmedeus
	# for linux build on mac
	GOOS=linux GOARCH=amd64 CC="/usr/local/bin/x86_64-linux-musl-gcc" CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" -o dist/osmedeus
	zip -9 -j dist/osmedeus-linux.zip dist/osmedeus
	rm -rf ./dist/osmedeus

release:
	go install
	rm -rf ./dist/* ~/myGit/premium-osmedeus-base/dist/* ~/org-osmedeus/osmedeus-base/dist/*
#	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o dist/osmedeus
#	zip -j dist/osmedeus-darwin.zip dist/osmedeus
#	rm -rf ./dist/osmedeus
	# for linux build on mac
	GOOS=linux GOARCH=amd64 CC="/usr/local/bin/x86_64-linux-musl-gcc" CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static"  -o dist/osmedeus
	zip -j dist/osmedeus-linux.zip dist/osmedeus
	rm -rf ./dist/osmedeus
	cp dist/* ~/myGit/premium-osmedeus-base/dist/
	cp dist/* ~/org-osmedeus/osmedeus-base/dist/
run:
	$(GO) $(GOFLAGS) run *.go

fmt:
	$(GO) $(GOFLAGS) fmt ./...; \
	echo "Done."

test:
	$(GO) $(GOFLAGS) test ./... -v%