HOSTNAME=localhost
NAMESPACE=backblaze
NAME=b2
BINARY=terraform-provider-${NAME}
VERSION=$(shell git describe --tags --abbrev=0 | cut -c2-)
OS_ARCH=${GOOS}_${GOARCH}

default: build

.PHONY: _pybindings deps format testacc clean build install docs

_pybindings:
ifeq ($(origin NOPYBINDINGS), undefined)
	$(MAKE) -C python-bindings $(MAKECMDGOALS)
else
	$(info Skipping python bindings (NOPYBINDINGS is defined))
endif


deps: _pybindings
	go mod download
	go get github.com/markbates/pkger/cmd/pkger

format: _pybindings
	go fmt ./...
	terraform fmt -recursive ./examples/

testacc: _pybindings
	chmod +rx python-bindings/dist/py-terraform-provider-b2
	TF_ACC=1 go test ./${NAME} -v -count 1 -parallel 4 -timeout 120m $(TESTARGS)

clean: _pybindings
	rm -rf pkged.go ${BINARY}

build: _pybindings
	pkger -include /python-bindings/dist/py-terraform-provider-b2
	go build -tags netgo -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

all: deps testacc build
