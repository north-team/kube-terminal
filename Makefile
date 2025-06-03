BINARY_NAME=helm-wrapper

GOPATH = $(shell go env GOPATH)

image ?= "registry.fit2cloud.com/north/kube-terminal"
branch ?= "dev-test-session"

LDFLAGS="-s -w"

# build docker image
build-docker:
	docker build -t ${image}:${branch} .

# build docker image
buildx-docker:
	#GOOS=linux GOARCH=arm64 go build -v -o ./kube-terminal
	docker buildx build --output "type=image,push=true" --platform linux/amd64,linux/arm64 --tag ${image}:${branch} .
