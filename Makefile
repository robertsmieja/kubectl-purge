export GO111MODULE=on

BINARY_NAME_BASE=kubectl-purge
ifeq ($(OS),Windows_NT)
	BINARY_NAME:=$(BINARY_NAME_BASE).exe
else
	BINARY_NAME:=$(BINARY_NAME_BASE)
endif

.PHONY: test
test:
	go test ./... -coverprofile cover.out

.PHONY: bin
bin: fmt vet
	go build -o bin/$(BINARY_NAME) github.com/robertsmieja/kubectl-purge/

.PHONY: install
install: fmt vet
	go install ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...
