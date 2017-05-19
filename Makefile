SHORT_NAME ?= k8s-claimer

include versioning.mk

LDFLAGS := -ldflags "-s -X main.version=${VERSION}"
REPO_PATH := github.com/deis/${SHORT_NAME}
DEV_ENV_IMAGE := quay.io/deis/go-dev:0.20.0
DEV_ENV_WORK_DIR := /go/src/${REPO_PATH}
DEV_ENV_PREFIX := docker run --rm -v ${CURDIR}:${DEV_ENV_WORK_DIR} -w ${DEV_ENV_WORK_DIR}
DEV_ENV_CMD := ${DEV_ENV_PREFIX} ${DEV_ENV_IMAGE}

DEIS_BINARY_NAME ?= ./deis
DEIS_APP_NAME ?= ${SHORT_NAME}

DIST_DIR := _dist
BINARY_NAME := k8s-claimer

build: build-binary docker-build
push: docker-push

bootstrap:
	${DEV_ENV_CMD} glide install

glideup:
	${DEV_ENV_CMD} glide up

build-binary:
	${DEV_ENV_PREFIX} -e CGO_ENABLED=0 ${DEV_ENV_IMAGE} go build -a -installsuffix cgo ${LDFLAGS} -o rootfs/bin/server

docker-build:
	docker build ${DOCKER_BUILD_FLAGS} -t ${IMAGE} rootfs
	docker tag ${IMAGE} ${MUTABLE_IMAGE}

test:
	${DEV_ENV_CMD} sh -c 'go test $$(glide nv)'

test-cover:
	${DEV_ENV_CMD} test-cover.sh

.PHONY: deploy
deploy:
	@${DEIS_BINARY_NAME} login ${DEIS_URL} --username="${DEIS_USERNAME}" --password="${DEIS_PASSWORD}"
	${DEIS_BINARY_NAME} pull ${IMAGE} -a ${DEIS_APP_NAME}

build-cli-cross:
	${DEV_ENV_CMD} gox -verbose ${LDFLAGS} -os="linux darwin " -arch="amd64 386" -output="${DIST_DIR}/${BINARY_NAME}-latest-{{.OS}}-{{.Arch}}" ./cli
	${DEV_ENV_CMD} gox -verbose ${LDFLAGS} -os="linux darwin" -arch="amd64 386" -output="${DIST_DIR}/${VERSION}/${BINARY_NAME}-${VERSION}-{{.OS}}-{{.Arch}}" ./cli

build-cli:
	go build ${LDFLAGS} -o k8s-claimer-cli ./cli

dist: build-cli-cross

install:
	helm upgrade k8s-claimer chart --install --namespace k8sclaimer --set image.org=${IMAGE_PREFIX},image.tag=${VERSION},${ARGS}

upgrade:
	helm upgrade k8s-claimer chart --namespace k8sclaimer --set image.org=${IMAGE_PREFIX},image.tag=${VERSION},${ARGS}

uninstall:
	helm delete k8s-claimer --purge
	