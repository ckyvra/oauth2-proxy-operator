IMG ?= ghcr.io/ckyvra/oauth2-proxy-operator:latest
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  = $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build
build:
	go build $(LDFLAGS) -o bin/manager main.go

.PHONY: run
run:
	go run main.go --leader-elect=false

.PHONY: docker-build
docker-build:
	docker build -t $(IMG) .

.PHONY: docker-push
docker-push:
	docker push $(IMG)

.PHONY: deploy
deploy:
	kustomize build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy:
	kustomize build config/default | kubectl delete -f -

SETUP_ENVTEST = $(shell go env GOPATH)/bin/setup-envtest

.PHONY: test
test: setup-envtest
	KUBEBUILDER_ASSETS="$(shell $(SETUP_ENVTEST) use -p path 1.31.x 2>/dev/null)" \
		go test ./controllers/ -v -count=1

.PHONY: test-all
test-all: setup-envtest
	KUBEBUILDER_ASSETS="$(shell $(SETUP_ENVTEST) use -p path 1.31.x 2>/dev/null)" \
		go test ./... -v -count=1

.PHONY: setup-envtest
setup-envtest:
	@command -v $(SETUP_ENVTEST) > /dev/null 2>&1 || \
		go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
	@$(SETUP_ENVTEST) use -p path 1.31.x > /dev/null 2>&1 || \
		$(SETUP_ENVTEST) use 1.31.x > /dev/null 2>&1

.PHONY: generate
generate:
	controller-gen object paths="./api/..."
	controller-gen crd paths="./api/..." output:crd:dir=config/crd

GOLANGCI_LINT = $(shell go env GOPATH)/bin/golangci-lint

.PHONY: lint
lint:
	@command -v $(GOLANGCI_LINT) > /dev/null 2>&1 || \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOLANGCI_LINT) run ./...
