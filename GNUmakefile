HOSTNAME=localhost
NAMESPACE=backblaze
NAME=b2
BINARY=terraform-provider-${NAME}
VERSION=$(shell git describe --tags --abbrev=0 | cut -c2-)
OS_ARCH=${GOOS}_${GOARCH}

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
	@go install github.com/markbates/pkger/cmd/pkger
	@go mod tidy
	@cd tools && go mod download
	@cd tools && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	@cd tools && go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
	@cd tools && go mod tidy

deps-check:
	@go mod tidy
	@cd tools && go mod tidy
	@git diff --exit-code -- go.mod go.sum tools/go.mod tools/go.sum || \
		(echo; echo "Unexpected difference in go.mod/go.sum files. Run 'go mod tidy' command or revert any go.mod/go.sum changes and commit."; exit 1)

format: _pybindings
	@go fmt ./...
	@terraform fmt -recursive ./examples/

lint: _pybindings
	@python scripts/check-gofmt.py '**/*.go' pkged.go
	@golangci-lint run ./...
	@python scripts/check-headers.py '**/*.go' pkged.go

testacc: _pybindings
	@chmod +rx python-bindings/dist/py-terraform-provider-b2
	TF_ACC=1 go test ./${NAME} -v -count 1 -parallel 4 -timeout 120m $(TESTARGS)

clean: _pybindings
	@rm -rf dist pkged.go ${BINARY}

build: _pybindings
	@pkger -include /python-bindings/dist/py-terraform-provider-b2
	go build -tags netgo -o ${BINARY}

install: build
	@mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

docs: build
	@tfplugindocs

docs-lint:
	@tfplugindocs validate

all: deps lint build testacc
