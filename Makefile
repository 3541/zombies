GO_SRC := $(wildcard **/*.go)
BUILD_TYPE ?= "debug"

.PHONY: all run exec clean

all: zombies

clean:
	rm -f ./zombies

run: zombies exec clean

exec:
	./zombies

zombies: $(GO_SRC)
	go build -tags $(BUILD_TYPE)
