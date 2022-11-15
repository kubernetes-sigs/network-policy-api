HOST_GOOS ?= $(shell go env GOOS)
HOST_GOARCH ?= $(shell go env GOARCH)

test:
	go test ./pkg/...

fmt:
	go fmt ./cmd/... ./pkg/...

vet:
	go vet ./cmd/... ./pkg/...

cyclonus:
	CGO_ENABLED=0 GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH) go build -o ./cmd/cyclonus/cyclonus ./cmd/cyclonus
	docker build -t docker.io/mfenwick100/cyclonus:$(HOST_GOOS)-$(HOST_GOARCH)-latest ./cmd/cyclonus
