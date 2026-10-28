HOSTNAME=localhost
NAMESPACE=backblaze
NAME=b2
BINARY=terraform-provider-${NAME}
VERSION=$(shell git describe --tags --abbrev=0 | cut -c2-)
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)
CHANGIE_VERSION=v1.24.0
CHANGIE=go run github.com/miniscruff/changie@$(CHANGIE_VERSION)
BASE_REF?=origin/master

default: build

.PHONY: _pybindings deps deps-check format lint vulncheck test testacc clean build install docs docs-lint ci changelog-check changelog-new release-check review-metrics

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
	@cd tools && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	@cd tools && go install golang.org/x/vuln/cmd/govulncheck@v1.3.0
	@cd tools && go mod tidy

deps-check:
	@go mod tidy
	@cd tools && go mod tidy
	@git diff --exit-code -- go.mod go.sum tools/go.mod tools/go.sum || \
		(echo; echo "Unexpected difference in go.mod/go.sum files. Run 'make deps' command or revert any go.mod/go.sum changes and commit."; exit 1)

format: _pybindings
	@golangci-lint fmt ./...
	@terraform fmt -recursive ./examples/

lint: _pybindings
	@python scripts/check-headers.py '**/*.go'
	@golangci-lint fmt --diff ./...
	@test -f b2/py-terraform-provider-b2 || touch b2/py-terraform-provider-b2 # required by go:embed in bindings.go
	@go vet ./...
	@golangci-lint run ./...

vulncheck:
	@test -f b2/py-terraform-provider-b2 || touch b2/py-terraform-provider-b2 # required by go:embed in bindings.go
	@govulncheck ./...

test:
	@test -f b2/py-terraform-provider-b2 || touch b2/py-terraform-provider-b2 # required by go:embed in bindings.go
	@go test ./... -v -count 1 -skip '^TestAcc' $(TESTARGS)

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
	# We clear B2 application key env variables to not include them in docs. DO NOT REMOVE IT!
	# Provider name is passed explicitly, otherwise tfplugindocs derives it from the directory name (breaks in git worktrees).
	@B2_APPLICATION_KEY_ID= B2_APPLICATION_KEY= tfplugindocs generate --provider-name ${BINARY}

docs-lint: build
	@tfplugindocs validate --provider-name ${BINARY}
	# We clear B2 application key env variables to not include them in docs. DO NOT REMOVE IT!
	@B2_APPLICATION_KEY_ID= B2_APPLICATION_KEY= tfplugindocs generate --provider-name ${BINARY}
	@git diff --exit-code -- docs/ || \
		(echo; echo "Unexpected difference in docs. Run 'make docs' command or revert any changes in the schema."; exit 1)

# Local mirror of the lint, test and build jobs in ci.yml plus changelog.yml (no acceptance tests). Run `make deps` first.
# Each step is its own make call: _pybindings forwards the goal names to python-bindings/GNUmakefile.
ci:
	@for target in deps-check lint test docs-lint changelog-check; do \
		echo "==> make $$target"; $(MAKE) --no-print-directory $$target || exit 1; \
	done
	@echo "==> make ci: all checks passed"

# A branch that changes anything outside the paths below must add a .changes/unreleased entry
# (or a release's .changes/vX.Y.Z.md). Existing entries must parse.
changelog-check:
	@git rev-parse --verify --quiet $(BASE_REF) > /dev/null || { echo "BASE_REF $(BASE_REF) not found; run 'git fetch origin' or pass BASE_REF=<ref>."; exit 1; }
	@if git diff --name-only $(BASE_REF)...HEAD -- . ':!*.md' ':!**/*.md' ':!*_test.go' ':!.changes/**' ':!intent/**' ':!.claude/**' ':!.github/ISSUE_TEMPLATE/**' ':!.editorconfig' | grep -q .; then \
		git diff --name-only --diff-filter=A $(BASE_REF)...HEAD -- '.changes/unreleased/*.yaml' '.changes/v*.md' | grep -q . \
		|| { echo "No changelog entry: run 'make changelog-new' (or label the PR -changelog)."; exit 1; }; \
	fi
	@if ls .changes/unreleased/*.yaml > /dev/null 2>&1; then $(CHANGIE) batch auto --dry-run > /dev/null; fi

changelog-new:
	@$(CHANGIE) new

# Holds a release while solvable security or quality findings are open (see RELEASE.md).
release-check:
	@scripts/release-check.sh

# Review health for merged PRs since SINCE (default: 90 days ago).
review-metrics:
	@scripts/review-metrics.sh $(SINCE)

all: deps lint build testacc
