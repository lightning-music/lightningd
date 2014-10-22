V_MAJOR=0
V_MINOR=1
V_PATCH=0
V=$(V_MAJOR).$(V_MINOR).$(V_PATCH)
PKG_DIR=lightning-$(V)
PKG=$(PKG_DIR).tar.gz

.PHONY: all pkg

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
