PACKAGE  = fhirbase
export GOPATH   = $(CURDIR)/.gopath
BASE     = $(GOPATH)/src/$(PACKAGE)
DATE    ?= $(shell date +%FT%T%z)
# VERSION ?= $(shell (cat $(BASE)/.version 2> /dev/null) || (echo 'nightly-\c' && git rev-parse --short HEAD 2> /dev/null)  | tr -d "\n")
VERSION := v0.1.0
GO      = go
GODOC   = godoc
GOFMT   = gofmt


.PHONY: darwin
darwin: lint fmt | $(BASE)
	GOOS=darwin GOARCH=amd64 $(GO) build \
	-v \
	-tags release \
	-ldflags '-X "main.Version=$(VERSION)" -X "main.BuildDate=$(DATE)"' \
	-o bin/$(PACKAGE)-darwin-amd64 *.go

.PHONY: linux
linux: lint fmt | $(BASE)
	GOOS=linux GOARCH=amd64 $(GO) build \
	-v \
	-tags release \
	-ldflags '-X "main.Version=$(VERSION)" -X "main.BuildDate=$(DATE)"' \
	-o bin/$(PACKAGE)-linux-amd64 *.go

.PHONY: windows
windows: lint fmt | $(BASE)
	GOOS=windows GOARCH=amd64 $(GO) build \
	-v \
	-tags release \
	-ldflags '-X "main.Version=$(VERSION)" -X "main.BuildDate=$(DATE)"' \
	-o bin/$(PACKAGE)-windows-amd64.exe *.go

.PHONY: windows-386
GOOS=windows GOARCH=386 $(GO) build \
	-v \
	-tags release \
	-ldflags '-X "main.Version=$(VERSION)" -X "main.BuildDate=$(DATE)"' \
	-o bin/$(PACKAGE)-windows-386.exe *.go

.PHONY: linux-386
linux-386: lint fmt | $(BASE)
	GOOS=linux GOARCH=386 $(GO) build \
	-v \
	-tags release \
	-ldflags '-X "main.Version=$(VERSION)" -X "main.BuildDate=$(DATE)"' \
	-o bin/$(PACKAGE)-linux-386 *.go

.PHONY: all
all: darwin linux windows linux-386 windows-386
# all: lint fmt | $(BASE)
# 	$(GO) build \
# 	-v \
# 	-tags release \
# 	-ldflags '-X "main.Version=$(VERSION)" -X "main.BuildDate=$(DATE)"' \
# 	-o bin/$(PACKAGE)$(BINSUFFIX) *.go

$(BASE):
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR) $@

# Tools

.PHONY: lint
lint: $(BASE) $(GOLINT)
	$Q cd $(BASE) && ret=0 && for pkg in $(PKGS); do \
	test -z "$$($(GOLINT) $$pkg | tee /dev/stderr)" || ret=1 ; \
	done ; exit $$ret

.PHONY: fmt
fmt:
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
	$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	done ; exit $$ret

.PHONY: clean
clean:
	go clean -modcache
	rm -rf bin .gopath vendor

.PHONY: tests
test: fmt lint
	go test $(ARGS)

# .PHONY: docker
# docker: Dockerfile bin/fhirbase-linux-amd64
# 	docker build . -t fhirbase/fhirbase:$(VERSION) -t fhirbase/fhirbase:latest && \
# 	docker push fhirbase/fhirbase:$(VERSION) && \
# 	docker push fhirbase/fhirbase:latest
