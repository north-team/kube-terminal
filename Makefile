BINARY_NAME=helm-wrapper

GOPATH = $(shell go env GOPATH)

image ?= "registry.fit2cloud.com/north/kube-terminal"
branch ?= "dev"

LDFLAGS="-s -w"

# build docker image
build-docker:
	docker build -t ${image}:${branch} .

# build docker image
buildx-docker:
	docker buildx build --output "type=image,push=true" --platform linux/amd64,linux/arm64 --tag ${image}:${branch} .
