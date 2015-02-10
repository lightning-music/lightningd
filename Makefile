PKG_DIR=lightning
PKG=$(PKG_DIR).tar.gz
README=README.md
PROG=lightningd
WWW_GIT=https://github.com/lightning/www.git

SRC=lightningd.go            \
    api.go                   \
    metro.go                 \
    note.go                  \
    pattern.go               \
    sequencer.go             \
    server.go

.PHONY: all pkg clean

# GOINSTALL := go install -ldflags -w -gcflags "-N -l"
GOINSTALL := go install -a

all .DEFAULT: $(PROG)

lightningd: $(SRC)
	go build $^

install:
	$(GOINSTALL)

$(PKG_DIR):
	mkdir $(PKG_DIR)

pkg: $(PKG_DIR) $(PROG)
	cp $(PROG) $(README) $(PKG_DIR)
	cd $(PKG_DIR) && git clone $(WWW_GIT)
	tar czf $(PKG) $(PKG_DIR)

clean:
	rm -rf *.tar.gz $(PROG) *.log *~ $(PKG_DIR)
