.PHONY = all clean

PLATFORMS := linux-arm6 linux-arm7 linux-amd64 linux-386 darwin-amd64
VERSION := $(shell git describe --always --tags --dirty="-dev-$$(git rev-parse --short HEAD)")
MAIN := ./cmd/lb112xutil

BUILDCMD := go build -o
ifneq ($(strip $(VERSION)),)
	BUILDCMD := go build -ldflags="-X 'main.Version=$(VERSION)'" -o
endif


TARGETS := $(foreach ku,$(PLATFORMS),lb112xutil-$(ku))
SUMS := SHA1SUM.txt SHA256SUM.txt

all: $(TARGETS) $(SUMS)
	@chmod +x $(TARGETS)

lb112xutil-linux-arm%:
	env GOOS=linux GOARCH=arm GOARM=$* $(BUILDCMD) $@ $(MAIN)

lb112xutil-linux-%:
	env GOOS=linux GOARCH=$* $(BUILDCMD) $@ $(MAIN)

lb112xutil-darwin-%:
	env GOOS=darwin GOARCH=$* $(BUILDCMD) $@ $(MAIN)

SHA%SUM.txt: $(TARGETS)
	shasum -a $* $(TARGETS) > $@

clean:
	@rm -fv $(TARGETS) $(SUMS)
