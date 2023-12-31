# BI Container Runtime

this project started as a fork of [teddyking/ns-process](https://github.com/teddyking/ns-process)

## Creating root filesystem

### Busybox

If you want minimal *nix env with busybox you can run:

```sh
export MAKEFLAGS=-j`nproc`
export ROOTFS="/tmp/rootfs"
make rootfs
```

when the config menu appears go to `Settings -> Build Options` and select `Build static binary`, then exit and save configuration.
This will download, build and install busybox in specified directory

and chroot if needed:
```
sudo chroot "$ROOTFS" /usr/bin/env -i   \
    HOME=/root                  \
    TERM="$TERM"                \
    PS1='(chroot) \u:\w\$ ' \
    PATH=/usr/bin:/usr/sbin     \
    /bin/sh --login
```

### Debootstrap

Alternatively, if you want full Debian environment you can use [debootstrap](https://wiki.debian.org/Debootstrap)

```sh
export ROOTFS=/tmp/rootfs
mkdir -pv $ROOTFS
mkdir -pv /tmp/dpkg.cache
debootstrap --cache-dir=/tmp/dpkg.cache bookworm $ROOTFS http://deb.debian.org/debian/
```

and chroot if needed:

```sh
sudo mount proc $ROOTFS/proc -t proc
sudo mount sysfs $ROOTFS/sys -t sysfs
sudo cp /etc/hosts $ROOTFS/etc/hosts
sudo cp /proc/mounts $ROOTFS/etc/mtab
sudo chroot $ROOTFS /bin/bash
```
