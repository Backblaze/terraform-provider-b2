SRC=$(shell find b2_terraform -name "*.py")
BINARY=dist/py-terraform-provider-b2
SPEC=py-terraform-provider-b2.spec

default: build

.PHONY: deps format testacc clean build

deps:
ifeq ($(origin VIRTUAL_ENV), undefined)
	pip install --user -r requirements.txt
else
	pip install -r requirements.txt
endif

format:
	black -Sl 100 b2_terraform

testacc: build

clean:
	rm -rf build dist py-terraform-provider-b2.egg-info

build:
	pyinstaller --log-level WARN --onefile ${SPEC}

install: build

all: deps build
