NAME = waybar-lyric

GO      ?= go
REVIVE  ?= revive
SRC_BIN ?= bin/$(NAME)
PREFIX  ?= /usr/local

BIN_FILE        = $(shell realpath -m "$(PREFIX)/bin/$(NAME)")
DOC_DIR         = $(shell realpath -m "$(PREFIX)/share/doc/$(NAME)")
DOC_FILE        = $(shell realpath -m "$(PREFIX)/share/doc/$(NAME)/README.md")
LICENSE_DIR     = $(shell realpath -m "$(PREFIX)/share/licenses/$(NAME)")
LICENSE_FILE    = $(shell realpath -m "$(PREFIX)/share/licenses/$(NAME)/LICENSE")
BASH_COMPLETION = $(shell realpath -m "$(PREFIX)/share/bash-completion/completions/$(NAME)")
ZSH_COMPLETION  = $(shell realpath -m "$(PREFIX)/share/zsh/site-functions/_$(NAME)")
FISH_COMPLETION = $(shell realpath -m "$(PREFIX)/share/fish/vendor_completions.d/$(NAME).fish")

-include Makefile.local

# Default target
.PHONY: all
all: build

# Build the Go binary
.PHONY: build
build:
	$(GO) build -trimpath -o $(SRC_BIN)

# Build the Go binary
.PHONY: test
test:
	$(GO) test -v -cover ./...
	$(REVIVE) -config revive.toml

# Clean up build artifacts
.PHONY: clean
clean:
	rm -f $(SRC_BIN)

.PHONY: install
install:
	install -Dsm755 $(SRC_BIN) "$(BIN_FILE)"
	install -Dm644  LICENSE    "$(LICENSE_FILE)"
	install -Dm644  README.md  "$(DOC_FILE)"

	$(SRC_BIN) _carapace bash | install -Dm644 /dev/stdin "$(BASH_COMPLETION)"
	$(SRC_BIN) _carapace zsh  | install -Dm644 /dev/stdin "$(ZSH_COMPLETION)"
	$(SRC_BIN) _carapace fish | install -Dm644 /dev/stdin "$(FISH_COMPLETION)"

.PHONY: uninstall
uninstall:
	@rm -vrf $(BIN_FILE) $(LICENSE_DIR) $(DOC_DIR) $(BASH_COMPLETION) $(ZSH_COMPLETION) $(FISH_COMPLETION)
