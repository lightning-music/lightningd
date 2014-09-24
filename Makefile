.PHONY: all

# GOINSTALL := go install -ldflags -w -gcflags "-N -l"
GOINSTALL := go install -a

all .DEFAULT:
	$(GOINSTALL)
