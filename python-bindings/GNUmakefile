NAME=py-terraform-provider-b2
DIR=b2_terraform
EGG_INFO=${NAME}.egg-info
SPEC=${NAME}.spec
UNAME_=$(shell uname)

default: build

.PHONY: deps format lint testacc clean build

deps:
	@pip install -r requirements-dev.txt

format:
	@black -Sl 100 ${DIR}

lint:
	@black --check -Sl 100 ${DIR}
	@flake8 --ignore=E501 ${DIR}
	@python ../scripts/check-headers.py '**/*.py'

testacc: build

clean:
	@rm -rf build dist ${EGG_INFO}

build:
	@pyinstaller ${SPEC}
ifeq ($(UNAME_), Linux)
	@mv -f dist/py-terraform-provider-b2 dist/py-terraform-provider-b2-linked
	@staticx --strip --loglevel INFO dist/py-terraform-provider-b2-linked dist/py-terraform-provider-b2
	@rm -f dist/py-terraform-provider-b2-linked
endif
ifneq ($(origin CI), undefined)
	@echo "version=$(subst refs/tags/v,,${GITHUB_REF})" > "${GITHUB_OUTPUT}"
endif

install: build

docs: build

all: deps lint build
