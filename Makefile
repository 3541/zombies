GO_SRC := $(wildcard **/*.go) $(wildcard *.go)
BUILD_TYPE ?= debug
HOST := $(shell uname | tr A-Z a-z)
TARGET ?= $(HOST)
ARCH ?= amd64
ifeq ($(BUILD_TYPE),release)
	FLAGS := -ldflags='-s -w'
else
	FLAGS := 
endif


.PHONY: all run exec clean

all: zombies

clean:
	rm -f ./zombies
	rm -rf ./build

run: zombies exec clean

exec:
	./zombies

ifeq ($(TARGET),$(HOST))
zombies: $(GO_SRC)
	GOOS=$(TARGET) go build -tags $(BUILD_TYPE) $(FLAGS)
else
zombies: $(GO_SRC)
	mkdir -p build
	cd build && xgo --targets=$(TARGET)/$(ARCH) $(FLAGS) ../
endif
