PACKAGE=github.com/salberternst/workspace
CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/bin
CLI_NAME=workspace

HEADHASH=$(shell git rev-parse HEAD)
TAG=$(shell git tag --points-at ${HEADHASH})
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION=${TAG}

ifeq ($(VERSION),)
	VERSION=${BRANCH}
endif

override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION}} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -extldflags "-static"

.PHONY: cli-linux-amd64
cli-linux-amd64:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${CLI_NAME} ./cmd/workspace/main.go
	zip 

.PHONY: cli-darwin-amd64
cli-darwin-amd64:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=darwin go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${CLI_NAME} ./cmd/workspace/main.go

.PHONY: cli-windows
cli-windows-amd64:
	mkdir -p bin
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${CLI_NAME}.exe ./cmd/workspace/main.go

bundle-linux-amd64: cli-linux-amd64
	mkdir -p bundles
	cd ${DIST_DIR}; zip "../bundles/${CLI_NAME}_linux_amd64_${VERSION}.zip" ${CLI_NAME}

bundle-darwin-amd64: cli-darwin-amd64
	mkdir -p bundles
	cd ${DIST_DIR}; zip "../bundles/${CLI_NAME}_darwin_amd64_${VERSION}.zip" ${CLI_NAME}

bundle-windows-amd64: cli-windows-amd64
	mkdir -p bundles
	cd ${DIST_DIR}; zip "../bundles/${CLI_NAME}_windows_amd64_${VERSION}.zip" ${CLI_NAME}.exe

make all: bundle-linux-amd64 bundle-darwin-amd64 bundle-windows-amd64