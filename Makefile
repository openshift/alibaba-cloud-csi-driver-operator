IMG ?= controller:latest

all: build

verify: fmt vet lint

build: alibaba-disk-csi-driver-operator

# Run go build to make alibaba-disk-csi-driver-operator
alibaba-disk-csi-driver-operator:
	go build -o bin/alibaba-disk-csi-driver-operator cmd/alibaba-disk-csi-driver-operator/main.go

# Run go fmt against code
.PHONY: fmt
fmt:
	go fmt ./...


# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

# Run golangci-lint against code
.PHONY: lint
lint: golangci-lint
	( GOLANGCI_LINT_CACHE=$(PROJECT_DIR)/.cache $(GOLANGCI_LINT) run --timeout 10m )

# Run go mod
.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

# Build the docker image
.PHONY: image
image:
	docker build -t ${IMG} .

# Push the docker image
.PHONY: push
push:
	docker push ${IMG}
