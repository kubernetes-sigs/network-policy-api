# TODO: create an actual image registry + image

# constants
# FIXME: update POLICY_ASSISTANT_IMAGE to the actual image name
POLICY_ASSISTANT_IMAGE			= policy-assistant

# variables
# FIXME: update IMAGE_REGISTRY to the actual image registry
IMAGE_REGISTRY      ?= docker.io
POLICY_ASSISTANT_TAG			?= latest

test:
	go test ./pkg/...

fmt:
	go fmt ./cmd/... ./pkg/...

vet:
	go vet ./cmd/... ./pkg/...

.PHONY: policy-assistant
policy-assistant: policy-assistant-binary

policy-assistant-binary:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/policy-assistant/policy-assistant ./cmd/policy-assistant

policy-assistant-image: policy-assistant-binary
	docker build -t $(IMAGE_REGISTRY)/$(POLICY_ASSISTANT_IMAGE):$(POLICY_ASSISTANT_TAG) ./cmd/policy-assistant
