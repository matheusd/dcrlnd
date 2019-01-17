# This is currently changed to use the local dir (.) as package instead of
# github.com/decred/dcrlnd so that makefile actions (such as running integration
# tests) are executed against the local dir (possibly with local changes)
# instead of trying to execute them against the official repo module's.
#PKG := github.com/decred/dcrlnd
PKG := .
FULLPKG := github.com/decred/dcrlnd
ESCPKG := github.com\/decred\/dcrlnd

DCRD_PKG := github.com/decred/dcrd
GOVERALLS_PKG := github.com/mattn/goveralls
LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint

GO_BIN := ${GOPATH}/bin
DCRD_BIN := $(GO_BIN)/dcrd
GOVERALLS_BIN := $(GO_BIN)/goveralls
LINT_BIN := $(GO_BIN)/golangci-lint

DCRD_DIR :=${GOPATH}/src/$(DCRD_PKG)

COMMIT := $(shell git describe --abbrev=40 --dirty)
LDFLAGS := -ldflags "-X $(FULLPKG)/build.Commit=$(COMMIT)"

DCRD_COMMIT := $(shell cat go.mod | \
		grep $(DCRD_PKG) | \
		head -n1 | \
		awk -F " " '{ print $$2 }' | \
		awk -F "/" '{ print $$1 }')

GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v
GOTEST := GO111MODULE=on go test -v

GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
GOLIST := go list $(PKG)/... | grep -v '/vendor/'
GOLISTCOVER := $(shell go list -f '{{.ImportPath}}' ./... | sed -e 's/^$(ESCPKG)/./')

RM := rm -f
CP := cp
MAKE := make
XARGS := xargs -L 1

include make/testing_flags.mk

DEV_TAGS := $(if ${tags},$(DEV_TAGS) ${tags},$(DEV_TAGS))

COVER = for dir in $(GOLISTCOVER); do \
		$(GOTEST) -tags="$(DEV_TAGS) $(LOG_TAGS)" \
			-covermode=count \
			-coverprofile=$$dir/profile.tmp $$dir; \
		\
		if [ $$? != 0 ] ; \
		then \
			exit 1; \
		fi ; \
		\
		if [ -f $$dir/profile.tmp ]; then \
			cat $$dir/profile.tmp | \
				tail -n +2 >> profile.cov; \
			$(RM) $$dir/profile.tmp; \
		fi \
	done

LINT = $(LINT_BIN) \
	run \
	--build-tags rpctest \
	--disable-all \
	--enable=gofmt \
	--enable=vet \
	--enable=gosimple \
	--enable=unconvert \
	--enable=ineffassign \
	--deadline=10m

GREEN := "\\033[0;32m"
NC := "\\033[0m"
define print
	echo $(GREEN)$1$(NC)
endef

default: scratch

all: scratch check install

# ============
# DEPENDENCIES
# ============

$(GOVERALLS_BIN):
	@$(call print, "Fetching goveralls.")
	go get -u $(GOVERALLS_PKG)

$(LINT_BIN):
	@$(call print, "Fetching golangci-lint")
	GO111MODULE=on go get -u $(LINT_PKG)

dcrd:
	@$(call print, "Installing dcrd $(DCRD_COMMIT).")
	GO111MODULE=on go get -v github.com/decred/dcrd/@$(DCRD_COMMIT)

# ============
# INSTALLATION
# ============

build:
	@$(call print, "Building debug dcrlnd and dcrlncli.")
	$(GOBUILD) -tags="$(DEV_TAGS)" -o dcrlnd-debug $(LDFLAGS) $(PKG)
	$(GOBUILD) -tags="$(DEV_TAGS)" -o dcrlncli-debug $(LDFLAGS) $(PKG)/cmd/dcrlncli

build-itest:
	@$(call print, "Building itest dcrlnd and dcrlncli.")
	$(GOBUILD) -tags="$(ITEST_TAGS)" -o dcrlnd-itest $(LDFLAGS) $(PKG)
	$(GOBUILD) -tags="$(ITEST_TAGS)" -o dcrlncli-itest $(LDFLAGS) $(PKG)/cmd/dcrlncli

install:
	@$(call print, "Installing dcrlnd and dcrlncli.")
	$(GOINSTALL) -tags="${tags}" $(LDFLAGS) $(PKG)
	$(GOINSTALL) -tags="${tags}" $(LDFLAGS) $(PKG)/cmd/dcrlncli

scratch: build


# =======
# TESTING
# =======

check: unit itest

itest-only:
	@$(call print, "Running integration tests.")
	$(ITEST)

itest: dcrd build-itest itest-only

unit-only:
	@$(call print, "Running unit tests.")
	$(UNIT)

unit: dcrd unit-only

unit-cover:
	@$(call print, "Running unit coverage tests.")
	echo "mode: count" > profile.cov
	$(COVER)

unit-race:
	@$(call print, "Running unit race tests.")
	env CGO_ENABLED=1 GORACE="history_size=7 halt_on_errors=1" $(UNIT_RACE)

goveralls: $(GOVERALLS_BIN)
	@$(call print, "Sending coverage report.")
	$(GOVERALLS_BIN) -coverprofile=profile.cov -service=travis-ci

travis-race: dcrd unit-race

travis-cover: dcrd unit-cover goveralls

# =============
# FLAKE HUNTING
# =============

flakehunter: build-itest
	@$(call print, "Flake hunting integration tests.")
	while [ $$? -eq 0 ]; do $(ITEST); done

flake-unit:
	@$(call print, "Flake hunting unit tests.")
	GOTRACEBACK=all $(UNIT) -count=1
	while [ $$? -eq 0 ]; do /bin/sh -c "GOTRACEBACK=all $(UNIT) -count=1"; done

# =========
# UTILITIES
# =========

fmt:
	@$(call print, "Formatting source.")
	gofmt -l -w -s $(GOFILES_NOVENDOR)

lint: $(LINT_BIN)
	@$(call print, "Linting source.")
	$(LINT)

list:
	@$(call print, "Listing commands.")
	@$(MAKE) -qp | \
		awk -F':' '/^[a-zA-Z0-9][^$$#\/\t=]*:([^=]|$$)/ {split($$1,A,/ /);for(i in A)print A[i]}' | \
		grep -v Makefile | \
		sort

rpc:
	@$(call print, "Compiling protos.")
	cd ./lnrpc; ./gen_protos.sh

clean:
	@$(call print, "Cleaning source.$(NC)")
	$(RM) ./dcrlnd-debug ./dcrlncli-debug
	$(RM) ./dcrlnd-itest ./dcrlncli-itest


.PHONY: all \
	dcrd \
	default \
	build \
	install \
	scratch \
	check \
	itest-only \
	itest \
	unit \
	unit-cover \
	unit-race \
	goveralls \
	travis-race \
	travis-cover \
	flakehunter \
	flake-unit \
	fmt \
	lint \
	list \
	rpc \
	clean
