PACKAGE  = fhirbase
export GOPATH   = $(CURDIR)/.gopath
BASE     = $(GOPATH)/src/$(PACKAGE)
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell (cat $(BASE)/.version 2> /dev/null) || (echo 'nightly-\c' && git rev-parse --short HEAD 2> /dev/null)  | tr -d "\n")

GO      = go
GODOC   = godoc
GOFMT   = gofmt

.PHONY: all
all: a_main-packr.go lint fmt | $(BASE)
	$Q cd $(BASE) && $(GO) build \
	-v \
	-tags release \
	-ldflags '-X "main.Version=$(VERSION)" -X "main.BuildDate=$(DATE)"' \
	-o bin/$(PACKAGE)$(BINSUFFIX) *.go

a_main-packr.go: $(GOPATH)/bin/packr
	rm -rfv $(GOPATH)/src/golang.org/x/tools/go/loader/testdata; \
	rm -rfv $(GOPATH)/src/golang.org/x/tools/cmd/fiximports/testdata; \
	rm -rfv $(GOPATH)/src/golang.org/x/tools/internal/lsp/testdata; \
	go clean -modcache; \
	$(GOPATH)/bin/packr -z

$(BASE):
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR) $@

# # install packr with go get because dep doesn't build binaries for us
$(GOPATH)/bin/packr:
	$(GO) get -u github.com/gobuffalo/packr/...

# Tools

.PHONY: packr
packr: $(GOPATH)/bin/packr
	rm -rfv $(GOPATH)/src/golang.org/x/tools/go/loader/testdata; \
	rm -rfv $(GOPATH)/src/golang.org/x/tools/cmd/fiximports/testdata; \
	rm -rfv $(GOPATH)/src/golang.org/x/tools/internal/lsp/testdata; \
	go clean -modcache; \
	$(GOPATH)/bin/packr -z

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
	rm -rf bin .gopath vendor *-packr.go

.PHONY: tests
test: fmt lint packr
	cd $(BASE) && go test $(ARGS)

.PHONY: docker
docker: Dockerfile bin/fhirbase-linux-amd64
	docker build . -t fhirbase/fhirbase:$(VERSION) && \
	docker push fhirbase/fhirbase:$(VERSION)
