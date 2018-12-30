#!/bin/bash

set -eo pipefail

if [[ -z "$VERSION" ]]; then
    echo "VERSION unset!"
    exit 1
fi

case $GOARCH in
    ''|amd64)   ARCH=amd64 ;;
    mips*)      ARCH=mips ;;
    *)          ARCH=$GOARCH ;;
esac

wolssh="wolssh-${GOARCH}"
wolssh="${wolssh%-}"

tmpdir="$(mktemp -d)"
trap "rm -rf $tmpdir" EXIT

echo "preparing data"
datadir=$tmpdir/data
install -Dm755 $wolssh $datadir/usr/bin/wolssh
install -Dm755 debian/wolssh.init $datadir/etc/init.d/wolssh
install -Dm644 debian/wolssh.default $datadir/etc/default/wolssh
tar --owner=root:0 --group=root:0 -czf $tmpdir/data.tar.gz -C $datadir .

echo "creating debian control"
debdir=$tmpdir/debian
mkdir $debdir
sed -e "s|@VERSION@|$VERSION|g; s|@ARCH@|$ARCH|g" \
    debian/control >$debdir/control
install -m755 debian/postinst $debdir/
install -m644 debian/conffiles $debdir/
tar --owner=root:0 --group=root:0 -czf $tmpdir/control.tar.gz -C $debdir .
echo '2.0' >$tmpdir/debian-binary

pkgname="wolssh-$VERSION-$ARCH.deb"
echo "creating $pkgname"
rm -f $pkgname
ar -rcs $pkgname $tmpdir/debian-binary $tmpdir/data.tar.gz $tmpdir/control.tar.gz
