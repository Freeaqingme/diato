export GOPATH:=$(shell pwd)

GO        ?= go
PKG       := ./src/diato/
# TODO: Do we also want to run with debug in production?
# the github.com/rjeczalik/notify prints a lot of debug
# stuff when this is set.
BUILDTAGS := debug
VERSION   ?= $(shell git describe --dirty --tags | sed 's/^v//' )

.PHONY: default
default: all

# find src/ -name .git -type d | sed -s 's/.git$//' | while read line; do echo -n "${line} " | sed 's/.\/src\///'; git -C $line rev-parse HEAD; done | sort > GLOCKFILE
.PHONY: deps
deps:
	go get -tags '$(BUILDTAGS)' -d -v diato/...
	go get github.com/robfig/glock
	git diff /dev/null GLOCKFILE | ./bin/glock apply .

.PHONY: diato
diato: deps binary

.PHONY: binary
binary: LDFLAGS += -X "main.buildTag=v$(VERSION)"
binary: LDFLAGS += -X "main.buildTime=$(shell date -u '+%Y-%m-%d %H:%M:%S UTC')"
binary:
	go install -tags '$(BUILDTAGS)' -ldflags '$(LDFLAGS)' diato

.PHONY: release
release: BUILDTAGS=release
release: diato

.PHONY: fmt
fmt:
	go fmt diato/...

.PHONY: all
all: fmt diato

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf pkg/
	rm -rf src/diato/assets/
	go clean -i -r diato

.PHONY: test
test:
	go test -tags '$(BUILDTAGS)' -ldflags '$(LDFLAGS)' diato/...
