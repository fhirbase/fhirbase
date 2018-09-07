export PACKAGE  = fhirbase
export GOPATH   = $(CURDIR)/.gopath
export BASE     = $(GOPATH)/src/$(PACKAGE)
export DATE    ?= $(shell date +%FT%T%z)
export VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
	cat $(CURDIR)/.version 2> /dev/null || echo v0)
export GO15VENDOREXPERIMENT=1

export GO      = go
export GODOC   = godoc
export GOFMT   = gofmt
export DEP     = dep

export V = 0
export Q = $(if $(filter 1,$V),,@)
export M = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: all
all: vendor a_fhirbase-packr.go lint fmt | $(BASE) ; $(info $(M) building executable…) @ ## Build program binary
	$Q cd $(BASE) && $(GO) build \
	-v \
	-tags release \
	-ldflags '-X $(PACKAGE)/cmd.Version=$(VERSION) -X $(PACKAGE)/cmd.BuildDate=$(DATE)' \
	-o bin/$(PACKAGE)$(BINSUFFIX) *.go

a_fhirbase-packr.go: $(GOPATH)/bin/packr
	$(GOPATH)/bin/packr -z

$(BASE):
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR) $@

vendor: | $(BASE)
	cd $(BASE) && $(DEP) ensure
	for dep in `cd $(BASE)/vendor && find * -type d -maxdepth 3 -mindepth 2`; do \
	echo "building $$dep"; \
	go install -v $(PACKAGE)/vendor/$$dep || echo "not good"; \
	done
	mkdir -p $(GOPATH)/pkg/`go env GOOS`_`go env GOARCH`
	cp -r $(GOPATH)/pkg/`go env GOOS`_`go env GOARCH`/$(PACKAGE)/vendor/* $(GOPATH)/pkg/`go env GOOS`_`go env GOARCH`
	touch $@

# # install packr with go get because dep doesn't build binaries for us
$(GOPATH)/bin/packr:
	$(GO) get -u github.com/gobuffalo/packr/...

# Tools

.PHONY: packr
packr: $(GOPATH)/bin/packr ; $(info $(M) running packr…)
	$(GOPATH)/bin/packr -z

.PHONY: lint
lint: vendor | $(BASE) $(GOLINT) ; $(info $(M) running golint…) @ ## Run golint
	$Q cd $(BASE) && ret=0 && for pkg in $(PKGS); do \
	test -z "$$($(GOLINT) $$pkg | tee /dev/stderr)" || ret=1 ; \
	done ; exit $$ret

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
	$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	done ; exit $$ret

.PHONY: clean
clean:
	rm -rf bin .gopath vendor *-packr.go

.PHONY: tests
test: fmt lint vendor packr
	cd $(BASE) && go test $(ARGS)
