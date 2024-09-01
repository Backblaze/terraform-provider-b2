HOSTNAME=localhost
NAMESPACE=backblaze
NAME=b2
BINARY=terraform-provider-${NAME}
VERSION=$(shell git describe --tags --abbrev=0 | cut -c2-)
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

default: build

.PHONY: _pybindings deps deps-check format lint testacc clean build install docs docs-lint

_pybindings:
ifeq ($(origin NOPYBINDINGS), undefined)
	@$(MAKE) -C python-bindings $(MAKECMDGOALS)
else
	$(info Skipping python bindings (NOPYBINDINGS is defined))
endif

deps: _pybindings
	@go mod download
	@go mod tidy
	@cd tools && go mod download
	@cd tools && go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
	@cd tools && go mod tidy

deps-check:
	@go mod tidy
	@cd tools && go mod tidy
	@git diff --exit-code -- go.mod go.sum tools/go.mod tools/go.sum || \
		(echo; echo "Unexpected difference in go.mod/go.sum files. Run 'make deps' command or revert any go.mod/go.sum changes and commit."; exit 1)

format: _pybindings
	@go fmt ./...
	@terraform fmt -recursive ./examples/

lint: _pybindings
	@python scripts/check-gofmt.py '**/*.go' pkged.go
	@python scripts/check-headers.py '**/*.go' pkged.go

testacc: _pybindings
	@cp python-bindings/dist/py-terraform-provider-b2 b2/
	@chmod +rx b2/py-terraform-provider-b2
	TF_ACC=1 go test ./${NAME} -v -count 1 -parallel 4 -timeout 120m $(TESTARGS)

clean: _pybindings
	@rm -rf dist b2/py-terraform-provider-b2 ${BINARY}

build: _pybindings
	@cp python-bindings/dist/py-terraform-provider-b2 b2/
	@go build -tags netgo -o ${BINARY}

install: build
	@mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	@mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

docs: build
	@tfplugindocs

docs-lint: build
	@tfplugindocs validate
	@tfplugindocs
	@git diff --exit-code -- docs/ || \
		(echo; echo "Unexpected difference in docs. Run 'make docs' command or revert any changes in the schema."; exit 1)

all: deps lint build testacc
