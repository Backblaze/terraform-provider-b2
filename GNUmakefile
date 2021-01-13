HOSTNAME=localhost
NAMESPACE=backblaze
NAME=b2
BINARY=terraform-provider-${NAME}
VERSION=0.1.1
OS_ARCH=${GOOS}_${GOARCH}
PY_BINDINGS=python-bindings

default: build

.PHONY: _pybindings deps format testacc clean build install

_pybindings:
	$(MAKE) -C ${PY_BINDINGS} $(MAKECMDGOALS)

deps: _pybindings
	go mod download
	go get github.com/markbates/pkger/cmd/pkger

format: _pybindings
	go fmt ./...

testacc: _pybindings
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

clean: _pybindings
	rm -rf pkged.go ${BINARY}

build: _pybindings
	pkger -include /python-bindings/dist/py-terraform-provider-b2
	go build -tags netgo -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

all: deps testacc build
