HOSTNAME=localhost
NAMESPACE=backblaze
NAME=b2
BINARY=terraform-provider-${NAME}
VERSION=0.1.1
OS_ARCH=${GOOS}_${GOARCH}

default: testacc

.PHONY: format testacc

format:
	go fmt ./...

deps:
	go mod download
	go get github.com/markbates/pkger/cmd/pkger

build:
	pkger -include /python-bindings/dist/py-terraform-provider-b2
	go build -tags netgo -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

all: format testacc install
