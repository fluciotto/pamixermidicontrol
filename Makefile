EXECUTABLES = go patch
K := $(foreach exec,$(EXECUTABLES), \
  $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

.PHONY: all pamixermidicontrol clean distclean

all: pamixermidicontrol

pamixermidicontrol:
	go mod vendor
	patch -p 0 < go-midi.patch
	go build

clean:
	rm pamixermidicontrol

distclean: clean
	rm -rf vendor
