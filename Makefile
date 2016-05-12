SHORT_NAME ?= k8s-claimer

include versioning.mk

LDFLAGS := "-s -X main.version=${VERSION}"

REPO_PATH := github.com/deis/${SHORT_NAME}
DEV_ENV_IMAGE := quay.io/deis/go-dev:0.9.1
DEV_ENV_WORK_DIR := /go/src/${REPO_PATH}
DEV_ENV_PREFIX := docker run --rm -e GO15VENDOREXPERIMENT=1 -v ${CURDIR}:${DEV_ENV_WORK_DIR} -w ${DEV_ENV_WORK_DIR}
DEV_ENV_CMD := ${DEV_ENV_PREFIX} ${DEV_ENV_IMAGE}

DEIS_BINARY_NAME ?= ./deis

bootstrap:
	${DEV_ENV_CMD} glide install

build:
	${DEV_ENV_PREFIX} -e CGO_ENABLED=0 ${DEV_ENV_IMAGE} go build -a -installsuffix cgo -ldflags ${LDFLAGS} -o rootfs/bin/boot

test:
	${DEV_ENV_CMD} sh -c 'go test $$(glide nv)'

docker-build:
	docker build --rm -t ${IMAGE} rootfs
	docker tag -f ${IMAGE} ${MUTABLE_IMAGE}

deploy-to-deis:
	${DEIS_BINARY_NAME} pull ${IMAGE} -a ${DEIS_APP_NAME}

build-cli-cross:
	${DEV_ENV_CMD} gox -output="cli/bin/${SHORT_NAME}-{{.OS}}-{{.Arch}}"

prep-bintray-json:
	@jq '.version.name |= "$(VERSION)"' _scripts/ci/bintray-template.json \
		> _scripts/ci/bintray-ci.json

build-cli:
	go build -o k8s-claimer-cli ./cli

dist: build-cli-cross
