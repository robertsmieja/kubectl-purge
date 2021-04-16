
export GO111MODULE=on

BINARY_NAME_BASE=kubectl-purge
ifeq ($(OS),Windows_NT)
	BINARY_NAME=$(BINARY_NAME_BASE).exe
else
	BINARY_NAME=$(BINARY_NAME_BASE)
endif

.PHONY: test
test:
	go test ./pkg/... ./cmd/... -coverprofile cover.out

.PHONY: bin
bin: fmt vet
	go build -o bin/$(BINARY_NAME) github.com/robertsmieja/kubectl-purge/cmd/plugin

.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet:
	go vet ./pkg/... ./cmd/...

.PHONY: kubernetes-deps
kubernetes-deps:
	go get k8s.io/client-go@v11.0.0
	go get k8s.io/api@kubernetes-1.14.0
	go get k8s.io/apimachinery@kubernetes-1.14.0
	go get k8s.io/cli-runtime@kubernetes-1.14.0
