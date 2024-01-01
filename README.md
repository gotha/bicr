# BI Container Runtime

## Build

```sh
make build
```

## Run

!Note: Make sure you have created rootfs first and ROOTFS env variable is set (see below).

Without any parameters bicr-run starts a shell inside the container

```sh
./build/bicr-run
```

you can run specific commands from the container:

```sh
./build/bicr-run /bin/env
```

You can test with the example web server

```sh
cp ./build/bicr-httpd-example ./build/rootfs/opt
PORT=9999 ./build/bicr-run /opt/bicr-httpd-example
```


### Creating root filesystem

```sh
export ROOTFS="$(pwd)/build/rootfs"
mkdir -pv $ROOTFS
```

Depending on the flavour of Linux you like, you can use either minimalistic BusyBox environment or Debian, Arch, Fedora. 

Here are some examples:

#### [busybox](https://busybox.net/)

##### From source code

This is the default one.

You need to have installed `gcc`, `g++`, `make` and what Debian calls `build-essentials`

```sh
export MAKEFLAGS=-j`nproc` && make rootfs
```

when the config menu appears go to `Settings -> Build Options` and select `Build static binary`, then exit and save configuration.

##### Get busybox out of docker image

```sh
docker create busybox
CID=$(docker container ls -a --format json | grep busybox | jq -r '.ID')
docker export $CID | tar -xf - -C $ROOTFS
```

#### [debootstrap](https://wiki.debian.org/Debootstrap)

```sh
debootstrap bookworm $ROOTFS http://deb.debian.org/debian/
```

#### [pacstrap](https://wiki.archlinux.org/title/Pacstrap) on ArchLinux, Manjaro, etc

```sh
pacstrap -K $ROOTFS base vim
```

#### DNF on Fedora, RHEL, etc

```sh
sudo dnf -y --releasever=39 --installroot=$ROOTFS \
      --repo=fedora --repo=updates --setopt=install_weak_deps=False install \
      passwd dnf fedora-release vim-minimal 
```

#### Chroot in rootfs

If you want to start a shell in the new root filesystem you can:

```sh
sudo chroot $ROOTFS /bin/sh
```

### Credits

this project started as a fork of [teddyking/ns-process](https://github.com/teddyking/ns-process)


## Extending and reusing rootfs with Overlay FS

After you build `$ROOTFS`:

```sh
export ROOTFS_BASE="$ROOTFS-base"
export CONTAINERFS="$(pwd)/build/rootfs-container"

mv $ROOTFS $ROOTFS_BASE
mkdir -pv $ROOTFS $CONTAINERFS
mount -t tmpfs tmpfs $CONTAINERFS
mkdir -pv $CONTAINERFS/{up,work}

# note that we are copying the binary inside the 'up' directory
cp ./build/bicr-httpd-example $CONTAINERFS/up/bin/httpd

mount -t overlay overlay -o lowerdir=$ROOTFS_BASE,upperdir=$CONTAINERFS/up/,workdir=$CONTAINERFS/work/ $ROOTFS

# run container
./build/bicr-run httpd
```

Overlay FS applies everything inside `$CONTAINERFS` on top of `$ROOTFS_BASE` and makes accessible via `$ROOTFS` so you can `chroot` in it as per usual.
Needless to say you can have many `$CONTAINERFS` on top of a single base rootfs.

If you change the base root filesystem or the container FS, you would need to remount so changes can be visible in $ROOTFS.
