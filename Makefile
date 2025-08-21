GO         ?= go
REVIVE     ?= revive
SRCBIN     ?= ./bin/waybar-lyric
PREFIX     ?= /usr/local

BINDIR     ?= $(PREFIX)/bin
DOCDIR     ?= $(PREFIX)/share/doc/waybar-lyric
LICENSEDIR ?= $(PREFIX)/share/licenses/waybar-lyric
BASHCOMPDIR?= $(PREFIX)/share/bash-completion/completions
FISHCOMPDIR?= $(PREFIX)/share/fish/vendor_completions.d
ZSHCOMPDIR ?= $(PREFIX)/share/zsh/site-functions

# Default target
.PHONY: all
all: build

# Build the Go binary
.PHONY: build
build:
	$(GO) build -v -trimpath -ldflags "-s -w" -o $(SRCBIN)

# Build the Go binary
.PHONY: test
test:
	$(GO) test -v -cover ./...
	$(REVIVE) -config revive.toml

# Clean up build artifacts
.PHONY: clean
clean:
	rm -f waybar-lyric

.PHONY: install
install:
	install -Dm755 $(SRCBIN) $(BINDIR)/waybar-lyric
	install -Dm644 LICENSE $(LICENSEDIR)/LICENSE
	install -Dm644 README.md $(DOCDIR)/README.md

	$(SRCBIN) _carapace bash | install -Dm644 /dev/stdin $(BASHCOMPDIR)/waybar-lyric
	$(SRCBIN) _carapace zsh  | install -Dm644 /dev/stdin $(ZSHCOMPDIR)/_waybar-lyric
	$(SRCBIN) _carapace fish | install -Dm644 /dev/stdin $(FISHCOMPDIR)/waybar-lyric.fish

.PHONY: uninstall
uninstall:
	rm -f $(BINDIR)/waybar-lyric
	rm -rf $(LICENSEDIR)
	rm -rf $(DOCDIR)
	rm -rf $(BASHCOMPDIR) $(ZSHCOMPDIR) $(FISHCOMPDIR)
