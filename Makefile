# Check needed executables
EXECUTABLES = go patch strip
K := $(foreach exec,$(EXECUTABLES), \
  $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

# Get git informations
git_tag := $(shell git describe --tags 2>/dev/null || echo "Unknown")
git_commit := $(shell git rev-parse --short HEAD || echo "Unknown")

# Get current date
current_date := $(shell date --utc +"%Y-%m-%d:T%H:%M:%SZ")

# Linker flags
ld_flags = '-X "github.com/fluciotto/pamixermidicontrol/src.commit=${git_commit}" \
	-X "github.com/fluciotto/pamixermidicontrol/src.version=${git_tag}" \
	-X "github.com/fluciotto/pamixermidicontrol/src.buildTime=${current_date}"'

.PHONY: all pamixermidicontrol clean distclean

all: pamixermidicontrol

pamixermidicontrol:
	go mod vendor
	patch -p 0 < go-midi.patch
	go build -ldflags=${ld_flags}
	strip $@

clean:
	rm pamixermidicontrol

distclean: clean
	rm -rf vendor
