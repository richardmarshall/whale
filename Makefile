RUNTIME_DIR ?= /var/run/whale
.PHONY: build rootfs clean

build: bin/whale bin/stage1 bin/stage2

bin/whale: $(shell find . -type f -name '*.go' 2>/dev/null)
	go build -o bin/whale cmd/whale/main.go

bin/stage1: $(shell find . -type f -name '*.go' 2>/dev/null)
	go build -o bin/stage1 cmd/stage1/main.go

bin/stage2: $(shell find . -type f -name '*.go' 2>/dev/null)
	go build -o bin/stage2 cmd/stage2/main.go

$(RUNTIME_DIR):
	mkdir -p $(RUNTIME_DIR)/rootfs
	mkdir -p $(RUNTIME_DIR)/containers

rootfs: $(RUNTIME_DIR) $(RUNTIME_DIR)/rootfs/alpine $(RUNTIME_DIR)/rootfs/debian $(RUNTIME_DIR)/rootfs/busybox

$(RUNTIME_DIR)/rootfs/alpine:
	docker pull alpine; docker save alpine | undocker -i -o $(RUNTIME_DIR)/rootfs/alpine

$(RUNTIME_DIR)/rootfs/debian:
	docker pull debian; docker save debian | undocker -i -o $(RUNTIME_DIR)/rootfs/debian

$(RUNTIME_DIR)/rootfs/busybox:
	docker pull busybox; docker save busybox | undocker -i -o $(RUNTIME_DIR)/rootfs/busybox

clean:
	rm bin/*

clean-rootfs:
	rm -rf $(RUNTIME_DIR)/rootfs/*

clean-containers:
	rm -rf $(RUNTIME_DIR)/containers/*

cleanall: clean clean-rootfs clean-containers
