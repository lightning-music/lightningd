PKG_DIR=lightning
PKG=$(PKG_DIR).tar.gz

.PHONY: all pkg clean

# GOINSTALL := go install -ldflags -w -gcflags "-N -l"
GOINSTALL := go install -a

all .DEFAULT: lightningd

lightningd:
	go build lightningd.go

install:
	$(GOINSTALL)

$(PKG_DIR):
	mkdir $(PKG_DIR)

pkg: $(PKG_DIR) lightningd
	cp lightningd README $(PKG_DIR)
	tar czf $(PKG) $(PKG_DIR)

clean:
	rm -rf *.tar.gz lightningd *.log lightning-* *~ $(PKG_DIR)
