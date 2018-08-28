PACKAGE  = fhirbase
GOPATH   = $(CURDIR)/.gopath
BASE     = $(GOPATH)/src/$(PACKAGE)
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
	cat $(CURDIR)/.version 2> /dev/null || echo v0)

GO      = go
GODOC   = godoc
GOFMT   = gofmt
GLIDE   = glide

V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: all
all: lint fmt vendor packr | $(BASE) ; $(info $(M) building executable…) @ ## Build program binary
	$Q cd $(BASE) && $(GO) build \
	-v \
	-tags release \
	-ldflags '-X $(PACKAGE)/cmd.Version=$(VERSION) -X $(PACKAGE)/cmd.BuildDate=$(DATE)' \
	-o bin/$(PACKAGE) *.go

$(BASE):
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR) $@

glide.lock: glide.yaml | $(BASE)
	cd $(BASE) && $(GLIDE) update
	@touch $@

vendor: glide.lock | $(BASE)
	cd $(BASE) && $(GLIDE) --quiet install
	@ln -sf . vendor/src
	@touch $@

# install packr with go get because glide doesn't build binaries for us
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
	cd $(BASE) && go test
