PROJECT_NAME=bicr
BUSYBOX_VERSION="1.36.1"

.PHONY: build test rootfs-empty rootfs clean clean-rootfs

build:
	@echo ">>> Building Application..."
	mkdir -pv ./build
	go build -o ./build/${PROJECT_NAME}-run ./cmd/run
	go build -o ./build/${PROJECT_NAME}-build ./cmd/build
	go build -o ./build/${PROJECT_NAME}-httpd-example ./cmd/httpd-example

test:
	@echo ">>> Running Tests..."
	go test -race -v ./...

busybox:
	mkdir -pv ./build/busybox
	wget -O ./build/busybox.tar.bz2 https://busybox.net/downloads/busybox-${BUSYBOX_VERSION}.tar.bz2 
	tar -C ./build/busybox -xjf ./build/busybox.tar.bz2
	cd ./build/busybox/busybox-${BUSYBOX_VERSION} && $(MAKE) menuconfig && $(MAKE)
	cp ./build/busybox/busybox-${BUSYBOX_VERSION}/busybox ${ROOTFS}
	${ROOTFS}/busybox --install ${ROOTFS}/bin

rootfs-empty:
	mkdir -pv ${ROOTFS}/{etc,dev,run,var,tools,proc,sys,root,home,tmp} ${ROOTFS}/usr/{bin,lib,lib64,sbin} ${ROOTFS}/var/{db}
	cd ${ROOTFS} && \
	  for i in bin lib sbin; do \
	    ln -sv usr/$$i ${ROOTFS}/$$i; \
	  done\

rootfs: rootfs-empty busybox
	cp ./etc/* ${ROOTFS}/etc/

clean:
	rm -rf ./build

install: build
	cp ./build/${PROJECT_NAME}-run /usr/local/bin
	cp ./build/${PROJECT_NAME}-build /usr/local/bin
