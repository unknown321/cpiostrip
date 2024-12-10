PRODUCT=cpiostrip
GOOS=linux
GOARCH=$(shell uname -m)
GOARM=
ifeq ($(GOARCH),x86_64)
	override GOARCH=amd64
endif

NAME=$(PRODUCT)-$(GOOS)-$(GOARCH)$(GOARM)

$(NAME):
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a \
		-ldflags "-w -s" \
		-trimpath \
		-o $(NAME)

build: $(NAME)

all:
	$(MAKE) build
	$(MAKE) GOARCH=arm GOARM=5 build

release: all

clean:
	rm -rfv $(PRODUCT)-* test/

.PHONY: release all build clean
.DEFAULT_GOAL := all
