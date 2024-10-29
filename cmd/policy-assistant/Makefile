# TODO: create an actual image registry + image

# constants
# FIXME: update POLA_IMAGE to the actual image name
POLA_IMAGE			= pola

# variables
# FIXME: update IMAGE_REGISTRY to the actual image registry
IMAGE_REGISTRY      ?= docker.io
POLA_TAG			?= latest

test:
	go test ./pkg/...

fmt:
	go fmt ./cmd/... ./pkg/...

vet:
	go vet ./cmd/... ./pkg/...

.PHONY: pola
pola: pola-binary

pola-binary:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/pola/pola ./cmd/pola

pola-image: pola-binary
	docker build -t $(IMAGE_REGISTRY)/$(POLA_IMAGE):$(POLA_TAG) ./cmd/pola
