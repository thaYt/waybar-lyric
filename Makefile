SRCBIN     ?= ./bin/waybar-lyric
PREFIX     ?= /usr/local
BINDIR     ?= $(PREFIX)/bin
LICENSEDIR ?= $(PREFIX)/share/licenses/waybar-lyric
DOCDIR     ?= $(PREFIX)/share/doc/waybar-lyric

# Default target
.PHONY: all
all: build

# Build the Go binary
.PHONY: build
build:
	go build -trimpath -ldflags "-s -w" -o $(SRCBIN)

# Clean up build artifacts
.PHONY: clean
clean:
	rm -f waybar-lyric

.PHONY: install
install: 
	install -Dm 755 $(SRCBIN) $(DESTDIR)$(BINDIR)/waybar-lyric
	install -Dm 644 LICENSE $(DESTDIR)$(LICENSEDIR)/LICENSE
	install -Dm 644 README.md $(DESTDIR)$(DOCDIR)/README.md

.PHONY: uninstall
uninstall:
	rm -f $(DESTDIR)$(BINDIR)/waybar-lyric
	rm -rf $(DESTDIR)$(LICENSEDIR)
	rm -rf $(DESTDIR)$(DOCDIR)
